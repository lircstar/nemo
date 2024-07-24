package network

import (
	"errors"
	"net"
	"sync"
	"time"

	"nemo/sys/log"
	"nemo/sys/pool"
)

type TCPServer struct {
	Addr            string
	MaxConnNum      int
	PendingWriteNum int
	NewAgent        func(Conn) Agent
	ln              net.Listener
	connPool        *pool.ObjectPool

	wgLn    sync.WaitGroup
	wgConns sync.WaitGroup

	// msg parser
	LenMsgLen    int
	MinMsgLen    int
	MaxMsgLen    int
	LittleEndian bool
	msgParser    *TcpMsgParser
}

func (server *TCPServer) Start() {
	server.init()
	go server.run()
}

func (server *TCPServer) init() {
	ln, err := net.Listen("tcp", server.Addr)
	if err != nil {
		log.Fatalf("%v", err)
	}

	if server.MaxConnNum <= 0 {
		server.MaxConnNum = 100
		log.Warnf("invalid MaxConnNum, reset to %v", server.MaxConnNum)
	}
	if server.PendingWriteNum <= 0 {
		server.PendingWriteNum = 100
		log.Warnf("invalid PendingWriteNum, reset to %v", server.PendingWriteNum)
	}
	if server.NewAgent == nil {
		log.Fatal("NewAgent must not be nil")
	}

	server.ln = ln

	// connection pool
	server.connPool = pool.NewObjectPool()

	// msg parser
	msgParser := newTcpMsgParser()
	msgParser.SetMsgLen(server.LenMsgLen, server.MinMsgLen, server.MaxMsgLen)
	msgParser.SetByteOrder(server.LittleEndian)
	server.msgParser = msgParser
}

func (server *TCPServer) newTCPConn(conn net.Conn) *TCPConn {
	if server.connPool.FreeCount() <= 1 {
		for i := 0; i < 128; i++ {
			tcpConn := newTCPConn(server.PendingWriteNum, server.msgParser)
			server.connPool.Create(tcpConn)
		}
	}

	tcpConn := server.connPool.Get().(*TCPConn)
	tcpConn.closeFlag.Store(false)
	tcpConn.conn = nil
	tcpConn.bindConn(conn)
	return tcpConn
}

func (server *TCPServer) delTCPConn(tcpConn *TCPConn) {
	if tcpConn != nil {
		tcpConn.Close()
		server.connPool.Free(tcpConn)
	}
}

func (server *TCPServer) run() {
	server.wgLn.Add(1)
	defer server.wgLn.Done()

	var tempDelay time.Duration
	for {
		conn, err := server.ln.Accept()
		if err != nil {
			if !errors.Is(err, net.ErrClosed) {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				log.Errorf("accept error: %v; retrying in %v", err, tempDelay)
				time.Sleep(tempDelay)
				continue
			}
			return
		}
		tempDelay = 0

		if server.connPool.UsedCount() >= server.MaxConnNum {
			_ = conn.Close()
			log.Debug("too many connections")
			continue
		}

		server.wgConns.Add(1)

		tcpConn := server.newTCPConn(conn)
		tcpConn.start()
		agent := server.NewAgent(tcpConn)
		agent.SetType(TYPE_AGENT_TCP)
		go func() {
			// routine
			agent.OnConnect()
			agent.Run(nil)

			// cleanup
			server.delTCPConn(tcpConn)
			agent.OnClose()

			server.wgConns.Done()
		}()
	}
}

func (server *TCPServer) Close() {
	_ = server.ln.Close()
	server.wgLn.Wait()

	server.connPool.Range(func(i any) {
		if i != nil {
			conn := i.(*TCPConn).conn
			if conn != nil {
				_ = conn.Close()
			}
		}
	})

	server.wgConns.Wait()

	server.connPool = nil
}
