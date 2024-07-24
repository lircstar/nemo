package network

import (
	"nemo/sys/util"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"nemo/sys/log"
)

type TCPClients struct {
	Addr            string
	ConnNum         int
	ConnectInterval time.Duration
	PendingWriteNum int
	AutoReconnect   bool
	NewAgent        func(*TCPConn) Agent
	conns           *util.SafeMap
	wg              sync.WaitGroup
	closeFlag       atomic.Bool

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

	client.conns = util.NewSafeMap(client.ConnNum)
	client.closeFlag.Store(false)

	// msg parser
	msgParser := newTcpMsgParser()
	msgParser.SetMsgLen(client.LenMsgLen, client.MinMsgLen, client.MaxMsgLen)
	msgParser.SetByteOrder(client.LittleEndian)
	client.msgParser = msgParser
}

func (client *TCPClients) dial() net.Conn {
	for {
		conn, err := net.Dial("tcp", client.Addr)
		if err == nil || client.closeFlag.Load() {
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

	if client.closeFlag.Load() {
		err := conn.Close()
		if err != nil {
			return
		}
		return
	}
	client.conns.Store(conn, struct{}{})
	tcpConn := newTCPConn(client.PendingWriteNum, client.msgParser)
	tcpConn.bindConn(conn)
	tcpConn.start()
	agent := client.NewAgent(tcpConn)
	agent.SetType(TYPE_CLIENT_TCP)
	agent.OnConnect()
	agent.Run(nil)

	// cleanup
	tcpConn.Close()
	client.conns.Delete(conn)
	conn = nil
	tcpConn = nil
	agent.OnClose()
	agent = nil

	if client.AutoReconnect {
		time.Sleep(client.ConnectInterval)
		goto reconnect
	}
}

func (client *TCPClients) Close() {

	client.closeFlag.Store(true)

	client.conns.Range(func(key any, value any) bool {
		conn := key.(net.Conn)
		err := conn.Close()
		if err != nil {
			log.Errorf("close connection error: %v", err)
		}
		return true
	})
	client.conns = nil
	client.wg.Wait()
}

func (client *TCPClients) IsIn(tcpConn *TCPConn) bool {
	ret := false
	client.conns.Range(func(key any, value any) bool {
		conn := key.(net.Conn)
		if conn == tcpConn.conn {
			ret = true
		}
		return true
	})
	return ret
}
