package network

import (
	"net"
	"time"

	"github.com/lircstar/nemo/sys/log"
)

type TCPClient struct {
	Addr            string
	ConnectInterval time.Duration
	PendingWriteNum int
	AutoReconnect   bool
	agent           Agent
	NewAgent        func(Conn) Agent
	conn            net.Conn
	closeFlag       bool
	connected       bool

	// msg parser
	LenMsgLen    int
	MinMsgLen    int
	MaxMsgLen    int
	LittleEndian bool
	msgParser    *TcpMsgParser
}

func (client *TCPClient) Start() {
	client.init()
	go client.connect()
}

func (client *TCPClient) init() {

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
	if client.conn != nil {
		log.Fatal("client is running")
	}

	client.closeFlag = false

	// msg parser
	msgParser := newTcpMsgParser()
	msgParser.SetMsgLen(client.LenMsgLen, client.MinMsgLen, client.MaxMsgLen)
	msgParser.SetByteOrder(client.LittleEndian)
	client.msgParser = msgParser
}

func (client *TCPClient) dial() net.Conn {
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

func (client *TCPClient) connect() {

reconnect:
	conn := client.dial()
	if conn == nil {
		return
	}

	if client.closeFlag {
		err := conn.Close()
		if err != nil {
			return
		}
		return
	}

	tcpConn := newTCPConn(client.PendingWriteNum, client.msgParser)
	tcpConn.bindConn(conn)
	tcpConn.start()
	client.connected = true
	client.agent = client.NewAgent(tcpConn)
	client.agent.SetType(TYPE_CLIENT_TCP)
	client.agent.OnConnect()
	client.agent.Run(nil)
	client.connected = false

	// cleanup
	tcpConn.Close()
	conn = nil
	tcpConn = nil
	client.agent.OnClose()
	client.agent = nil

	if client.AutoReconnect {
		time.Sleep(client.ConnectInterval)
		goto reconnect
	}
}

func (client *TCPClient) Send(msg any) bool {
	return client.agent.SendMessage(msg)
}

func (client *TCPClient) Close() {
	client.closeFlag = true
	_ = client.conn.Close()
	client.conn = nil
}

func (client *TCPClient) GetType() uint {
	return TYPE_CLIENT_TCP
}

func (client *TCPClient) GetAddress() string {
	return client.Addr
}

func (client *TCPClient) GetConnected() bool {
	return client.connected
}

func (client *TCPClient) GetAgent() Agent {
	return client.agent
}
