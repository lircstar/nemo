package network

import (
	"fmt"
	"github.com/gorilla/websocket"
	"time"

	"github.com/lircstar/nemo/sys/log"
)

type WSClient struct {
	Addr             string
	ConnectInterval  time.Duration
	PendingWriteNum  int
	MaxMsgLen        int
	LittleEndian     bool
	HandshakeTimeout time.Duration
	AutoReconnect    bool
	agent            Agent
	NewAgent         func(Conn) Agent
	dialer           websocket.Dialer
	conn             *websocket.Conn
	closeFlag        bool
	connected        bool
}

func (client *WSClient) Start() {
	client.init()
	go client.connect()
}

func (client *WSClient) init() {

	if client.ConnectInterval <= 0 {
		client.ConnectInterval = 3 * time.Second
		log.Infof("invalid ConnectInterval, reset to %v", client.ConnectInterval)
	}
	if client.PendingWriteNum <= 0 {
		client.PendingWriteNum = 100
		log.Infof("invalid PendingWriteNum, reset to %v", client.PendingWriteNum)
	}
	if client.MaxMsgLen <= 0 {
		client.MaxMsgLen = 4096
		log.Infof("invalid MaxMsgLen, reset to %v", client.MaxMsgLen)
	}
	if client.HandshakeTimeout <= 0 {
		client.HandshakeTimeout = 10 * time.Second
		log.Infof("invalid HandshakeTimeout, reset to %v", client.HandshakeTimeout)
	}
	if client.NewAgent == nil {
		log.Fatal("NewAgent must not be nil")
	}
	if client.conn != nil {
		log.Fatal("client is running")
	}

	client.closeFlag = false

	client.dialer = websocket.Dialer{
		HandshakeTimeout: client.HandshakeTimeout,
	}
}

func (client *WSClient) dial() *websocket.Conn {
	for {
		conn, _, err := client.dialer.Dial(client.Addr, nil)
		if err == nil || client.closeFlag {
			fmt.Printf("connect to %v success\n", client.Addr)
			return conn
		}

		log.Errorf("connect to %v error: %v", client.Addr, err)
		time.Sleep(client.ConnectInterval)
		continue
	}
}

func (client *WSClient) connect() {

reconnect:
	conn := client.dial()
	if conn == nil {
		return
	}
	conn.SetReadLimit(int64(client.MaxMsgLen))

	if client.closeFlag {
		err := conn.Close()
		if err != nil {
			return
		}
		return
	}

	wsConn := newWSConn(conn, client.PendingWriteNum, client.MaxMsgLen)
	wsConn.conn = conn
	wsConn.start()
	client.agent = client.NewAgent(wsConn)
	client.agent.SetType(TYPE_CLIENT_WEBSOCKET)
	client.agent.OnConnect()
	client.agent.Run(nil)
	client.connected = false

	// cleanup
	wsConn.Close()
	conn = nil
	wsConn = nil
	client.agent.OnClose()
	client.agent = nil

	if client.AutoReconnect {
		time.Sleep(client.ConnectInterval)
		goto reconnect
	}
}

func (client *WSClient) Send(msg any) bool {
	return client.agent.SendMessage(msg)
}

func (client *WSClient) Close() {
	client.closeFlag = true
	if client.conn != nil {
		_ = client.conn.Close()
		client.conn = nil
	}
}

func (client *WSClient) GetType() uint {
	return TYPE_CLIENT_WEBSOCKET
}

func (client *WSClient) GetAddress() string {
	return client.Addr
}

func (client *WSClient) GetConnected() bool {
	return client.connected
}

func (client *WSClient) GetAgent() Agent {
	return client.agent
}
