package network

import (
	"nemo/sys/log"
	"nemo/sys/pool"
	"net"
	"sync"
	"time"
)

type UDPServer struct {
	Addr       string
	MaxConnNum int
	NewAgent   func(conn Conn) Agent
	ln         *net.UDPConn
	agents     AgentSet
	connPool   *pool.ObjectPool

	wgLn    sync.WaitGroup
	lkAgent sync.RWMutex

	// msg
	MinMsgLen int
	MaxMsgLen int

	LittleEndian bool
	msgParser    *UdpMsgParser

	timeEvent chan Conn
	running   bool // is server running?
}

func (server *UDPServer) Start(addr string) {
	server.Addr = addr
	server.init()
	go server.run()
}

func (server *UDPServer) init() {

	server.running = false

	if server.MaxConnNum <= 0 {
		server.MaxConnNum = 100
		log.Warnf("invalid UDP Server MaxConnNum, reset to %v", server.MaxConnNum)
	}

	server.agents = make(AgentSet, server.MaxConnNum)

	// connection pool
	server.connPool = pool.NewObjectPool()

	// msg parser
	msgParser := newUdpMsgParser()
	msgParser.SetMsgLen(server.MinMsgLen, server.MaxMsgLen)
	msgParser.SetByteOrder(server.LittleEndian)
	server.msgParser = msgParser

	server.timeEvent = make(chan Conn, 1024)
}

func (server *UDPServer) run() {
	server.wgLn.Add(1)
	defer server.wgLn.Done()

	addr, err := net.ResolveUDPAddr("udp", server.Addr)
	if err != nil {
		log.Errorf("udp address error %s; %v", server.Addr, err.Error())
		return
	}
	//addr := &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 9981}

	server.ln, err = net.ListenUDP("udp", addr)
	if err != nil {
		log.Errorf("udp listen error; %v", err.Error())
		return
	}

	if len(server.agents) >= server.MaxConnNum {
		log.Debug("udp server too many connections")
	}

	log.Infof("# UDP server started. %s", server.Addr)

	go server.accept()
	server.running = true
	go server.goRun()
}

func (server *UDPServer) accept() {
	recvBuff := make([]byte, server.MaxMsgLen)
	for {
		n, remoteAddr, err := server.ln.ReadFromUDP(recvBuff)
		if err != nil {
			log.Warnf("failed to udp read; err:%v", err.Error())
			break
		}

		if n > 0 && n >= server.MinMsgLen {
			go func() { // adjust go function to improve speed.
				agent := server.getAgent(remoteAddr)
				if agent != nil {
					agent.Run(recvBuff[:n])
				}
			}()
		}
	}

}

func (server *UDPServer) createConn() *UDPConn {
	if server.connPool.FreeCount() <= 1 {
		for i := 0; i < 128; i++ {
			conn := newUDPConn(server.msgParser)
			server.connPool.Create(conn)
		}
	}
	return server.connPool.Get().(*UDPConn)
}

func (server *UDPServer) getAgent(addr *net.UDPAddr) Agent {
	key := newConnTrackKey(addr)
	server.lkAgent.Lock()
	agent := server.agents[*key]
	server.lkAgent.Unlock()

	if agent == nil {
		conn := server.createConn()
		conn.timeEvent = server.timeEvent
		conn.closeFlag = false
		conn.conn = server.ln
		conn.remote = addr
		conn.key = key
		agent = server.NewAgent(conn)
		conn.agent = agent
		server.lkAgent.Lock()
		server.agents[*key] = agent
		server.lkAgent.Unlock()
		agent.OnConnect()
	}
	return agent
}

func (server *UDPServer) delConn(conn *UDPConn) {
	if conn != nil {
		if server.connPool != nil {
			server.connPool.Free(conn)
		}
	}
}

func (server *UDPServer) Close() {
	_ = server.ln.Close()
	server.wgLn.Wait()

	server.running = false
	// connection pool
	server.connPool.Range(func(i interface{}) {
		if i != nil {
			conn := i.(*UDPConn).conn
			if conn != nil {
				_ = conn.Close()
			}
		}
	})
	server.connPool = nil

	for _, agent := range server.agents {
		agent.Close()
	}
	server.agents = nil
}

func (server *UDPServer) goRun() {

	for server.running {
		select {
		case conn := <-server.timeEvent:
			udpConn := conn.(*UDPConn)
			server.delConn(udpConn)

			if udpConn.agent != nil {
				server.lkAgent.Lock()
				delete(server.agents, *udpConn.key)
				server.lkAgent.Unlock()
				udpConn.agent.OnClose()
			}
		case <-time.After(time.Second * 60):
		}
	}
}
