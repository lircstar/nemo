package network

import (
	"net"
	"time"

	"github.com/lircstar/nemo/sys/log"
)

type UDPClient struct {
	Addr            string
	ConnectInterval time.Duration
	AutoReconnect   bool
	running         bool

	conn     net.Conn
	agent    Agent
	NewAgent func(Conn) Agent

	idleTime int64
	TimeOut  int

	// msg
	MinMsgLen    int
	MaxMsgLen    int
	LittleEndian bool
	msgParser    *UdpMsgParser
}

func (client *UDPClient) Start() {
	client.init()
	go client.doConnect(client.Addr)
}

func (client *UDPClient) init() {
	if client.ConnectInterval <= 0 {
		client.ConnectInterval = 3 * time.Second
		log.Warnf("invalid ConnectInterval, reset to %v", client.ConnectInterval)
	}

	client.running = false

	// msg parser
	client.msgParser = newUdpMsgParser()
	client.msgParser.SetMsgLen(client.MinMsgLen, client.MaxMsgLen)
	client.msgParser.SetByteOrder(client.LittleEndian)
}

func (client *UDPClient) doConnect(remoteAddr string) {
	client.Addr = remoteAddr
	addr, err := net.ResolveUDPAddr("udp", client.Addr)
	if err != nil {
		log.Errorf("Failed to resolve udp address. addr : %s;  %v", client.Addr, err.Error())
		return
	}

reconnect:
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Errorf("Failed to dial udp")
		time.Sleep(client.ConnectInterval)
		return
	}

	udpConn := newUDPConn(client.msgParser)
	udpConn.conn = conn
	client.conn = conn
	client.agent = client.NewAgent(udpConn)
	client.agent.SetType(TYPE_CLIENT_UDP)
	client.running = true
	client.agent.OnConnect()
	client.idleTime = time.Now().Unix()
	go client.goRun()
	client.recv(udpConn)

	// cleanup
	client.running = false
	client.agent.OnClose()
	if udpConn.IsClosed() {
		if !client.AutoReconnect {
			return
		}
	} else {
		client.Close()
	}

	if client.AutoReconnect {
		time.Sleep(client.ConnectInterval)
		goto reconnect
	}
}

func (client *UDPClient) Close() {
	if client.agent != nil {
		client.conn.Close()
		client.agent.Close()
	}
}

func (client *UDPClient) Send(msg any) bool {
	if client.agent != nil {
		return client.agent.SendMessage(msg)
	}
	return false
}

func (client *UDPClient) recv(conn *UDPConn) {
	for !conn.IsClosed() {
		data, err := conn.ReadMsg()
		if err != nil {
			log.Warnf("failed to udp read; err:%v", err.Error())
			break
		}

		n := len(data)
		if n > 0 && n >= client.MinMsgLen {
			client.agent.Run(data)
			client.idleTime = time.Now().Unix()
		}
	}
}

func (client *UDPClient) goRun() {
	for client.running {
		select {
		case <-time.After(time.Second * 10):
			if time.Now().Unix()-client.idleTime > int64(client.TimeOut) {
				client.Close()
			}
		}
	}
}

func (client *UDPClient) GetType() uint {
	return TYPE_CLIENT_UDP
}

func (client *UDPClient) GetAddress() string {
	return client.Addr
}

func (client *UDPClient) GetConnected() bool {
	return true
}

func (client *UDPClient) GetAgent() Agent {
	return client.agent
}
