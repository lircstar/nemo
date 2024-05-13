package network

import (
	"net"
	"sync"
	"time"

	"nemo/sys/log"
)

type TCPClients struct {
	sync.Mutex
	Addr            string
	ConnNum         int
	ConnectInterval time.Duration
	PendingWriteNum int
	AutoReconnect   bool
	NewAgent        func(*TCPConn) Agent
	conns           ConnSet
	wg              sync.WaitGroup
	closeFlag       bool

	// msg parser
	LenMsgLen    int
	MinMsgLen    int
	MaxMsgLen    int
	LittleEndian bool
	msgParser    *TcpMsgParser
}

func (client *TCPClients) Start() {
	client.init()
	for i := 0; i < client.ConnNum; i++ {
		client.wg.Add(1)
		go client.connect()
	}
}

func (client *TCPClients) init() {
	client.Lock()
	defer client.Unlock()

	if client.ConnNum <= 0 {
		client.ConnNum = 1
		log.Warnf("invalid ConnNum, reset to %v", client.ConnNum)
	}
	if client.ConnectInterval <= 0 {
		client.ConnectInterval = 3 * time.Second
		log.Warnf("invalid ConnectInterval, reset to %v", client.ConnectInterval)
	}

	if client.PendingWriteNum <= 0 {
		client.PendingWriteNum = 100
		log.Warnf("invalid PendingWriteNum, reset to %v", client.PendingWriteNum)
	}
	if client.NewAgent == nil {
		log.Fatal("NewAgent must not be nil")
	}
	if client.conns != nil {
		log.Fatal("client is running")
	}

	client.conns = make(ConnSet)
	client.closeFlag = false

	// msg parser
	msgParser := newTcpMsgParser()
	msgParser.SetMsgLen(client.LenMsgLen, client.MinMsgLen, client.MaxMsgLen)
	msgParser.SetByteOrder(client.LittleEndian)
	client.msgParser = msgParser
}

func (client *TCPClients) dial() net.Conn {
	for {
		conn, err := net.Dial("tcp", client.Addr)
		if err == nil || client.closeFlag {
			return conn
		}

		log.Errorf("connect to %v error: %v", client.Addr, err)
		time.Sleep(client.ConnectInterval)
		continue
	}
}

func (client *TCPClients) connect() {
	defer client.wg.Done()

reconnect:
	conn := client.dial()
	if conn == nil {
		return
	}

	client.Lock()
	if client.closeFlag {
		client.Unlock()
		err := conn.Close()
		if err != nil {
			return
		}
		return
	}
	client.conns[conn] = struct{}{}
	client.Unlock()

	tcpConn := newTCPConn(client.PendingWriteNum, client.msgParser)
	tcpConn.bindConn(conn)
	tcpConn.start()
	agent := client.NewAgent(tcpConn)
	agent.OnConnect()
	agent.Run(nil)

	// cleanup
	tcpConn.Close()
	client.Lock()
	delete(client.conns, conn)
	conn = nil
	tcpConn = nil
	client.Unlock()
	agent.OnClose()
	agent = nil

	if client.AutoReconnect {
		time.Sleep(client.ConnectInterval)
		goto reconnect
	}
}

func (client *TCPClients) Close() {
	client.Lock()
	client.closeFlag = true
	for conn := range client.conns {
		err := conn.Close()
		if err != nil {
			continue
		}
	}
	client.conns = nil
	client.Unlock()

	client.wg.Wait()
}

func (client *TCPClients) IsIn(tcpConn *TCPConn) bool {
	client.Lock()
	defer client.Unlock()
	for conn := range client.conns {
		if conn == tcpConn.conn {
			return true
		}
	}
	return false
}
