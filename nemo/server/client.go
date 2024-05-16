package server

import (
	"fmt"
	"github.com/lircstar/nemo/nemo/conf"
	"github.com/lircstar/nemo/nemo/network"
	"github.com/lircstar/nemo/nemo/network/json"
	protobuf "github.com/lircstar/nemo/nemo/network/proto"
	"math"
	"time"
)

//-------------------------------------------------------------------------------------
// Connection manager to connect other server.
//-------------------------------------------------------------------------------------

//-------------------------------------------------------------------------------------
// Connect to a TCP server.

type TcpClientWrapper struct {
	network.TCPClient
}

// Connect Create a client and connect to a TCP server.
func (client *TcpClientWrapper) Connect(addr string) network.Client {
	client.Addr = addr
	client.AutoReconnect = conf.Reconnect
	client.ConnectInterval = 3 * time.Second
	client.PendingWriteNum = conf.PendingWriteNum
	client.LenMsgLen = conf.LenMsgLen
	client.MaxMsgLen = math.MaxInt32
	client.LittleEndian = conf.LittleEndian
	client.NewAgent = newClientAgent
	// If have no processor create by server, create it by itself.
	if processor == nil {
		processor = protobuf.NewProcessor()
	}

	client.Start()
	return client
}

func newClientAgent(conn network.Conn) network.Agent {
	a := new(Agent)
	a.conn = conn
	a.idleTime = time.Now().Unix()
	return a
}

//-------------------------------------------------------------------------------------
// Connect to a WebSocket server.

type WsClientWrapper struct {
	network.WSClient
}

func (client *WsClientWrapper) Connect(addr string) network.Client {
	client.Addr = addr
	client.AutoReconnect = conf.Reconnect
	client.ConnectInterval = 3 * time.Second
	client.PendingWriteNum = conf.PendingWriteNum
	client.MaxMsgLen = math.MaxInt32
	client.LittleEndian = conf.LittleEndian
	client.NewAgent = newClientAgent
	// If have no processor create by server, create it by itself.
	processor = json.NewProcessor()
	client.Start()
	return client
}

//-------------------------------------------------------------------------------------
// Connect to a UDP server.

type UdpClientWrapper struct {
	network.UDPClient
}

func (client *UdpClientWrapper) Connect(addr string) {
	client.Addr = addr
	client.TimeOut = conf.UdpTimeout
	client.MinMsgLen = conf.UdpMinMsgLen
	client.MaxMsgLen = conf.UdpMaxMsgLen
	client.LittleEndian = conf.LittleEndian
	client.AutoReconnect = conf.UdpReconnect
	client.ConnectInterval = conf.UdpConnectInterval
	client.NewAgent = newUdpClientAgent
	// If have no processor create by server, create it by itself.
	if processor == nil {
		fmt.Println("No processor found, use default protobuf.")
		processor = protobuf.NewProcessor()
	}
	client.Start()
}

func newUdpClientAgent(conn network.Conn) network.Agent {
	a := new(UdpAgent)
	a.conn = conn
	a.idleTime = time.Now().Unix()
	return a
}
