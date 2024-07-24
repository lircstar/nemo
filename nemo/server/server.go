package server

import (
	"nemo/nemo/conf"
	"nemo/nemo/network"
	"nemo/nemo/network/json"
	protobuf "nemo/nemo/network/proto"
	"nemo/sys/log"
	"nemo/sys/pool"
)

// -------------------------------------------------------------------------------------
// Tcp server.

type TcpServerWrapper struct {
	server *network.TCPServer
}

func (tcp *TcpServerWrapper) GetAddr() string {
	return tcp.server.Addr
}

func (tcp *TcpServerWrapper) GetType() uint {
	return TYPE_SERVER_TCP
}

func (tcp *TcpServerWrapper) Start() {

	config := conf.GetTCP()
	if len(config.Addr) == 0 {
		log.Error("ip adrress of server cannot be zero.")
		return
	}

	tcp.server = new(network.TCPServer)
	tcp.server.Addr = config.Addr
	tcp.server.MaxConnNum = config.MaxConnNum
	tcp.server.MaxConnNum = config.MinMsgLen
	tcp.server.MaxMsgLen = config.MaxMsgLen
	tcp.server.PendingWriteNum = 100
	tcp.server.NewAgent = newAgent
	tcp.server.LenMsgLen = config.LenMsgLen
	tcp.server.LittleEndian = LittleEndian

	if processor == nil {
		processor = protobuf.NewProcessor()
	}

	if tcp.server != nil {
		tcp.server.Start()
	}

	if onInitCallback != nil {
		onInitCallback()
	}

}

func (tcp *TcpServerWrapper) Stop() {
	if tcp.server != nil {
		tcp.server.Close()
	}
}

// -------------------------------------------------------------------------------------
// Websocket server.

type WsServerWrapper struct {
	server *network.WSServer
}

func (ws *WsServerWrapper) GetType() uint {
	return TYPE_SEVER_WEBSOCKET
}

func (ws *WsServerWrapper) GetAddr() string {
	return ws.server.Addr
}

func (ws *WsServerWrapper) Start() {
	config := conf.GetWSS()
	if len(config.Addr) == 0 {
		log.Error("adrress of server cannot be zero.")
		return
	}

	ws.server = new(network.WSServer)
	ws.server.Addr = config.Addr
	ws.server.MaxConnNum = config.MaxConnNum
	ws.server.MaxMsgLen = config.MaxMsgLen
	ws.server.PendingWriteNum = config.PendingWriteNum
	ws.server.HTTPTimeout = config.HTTPTimeout
	ws.server.CertFile = config.CertFile
	ws.server.KeyFile = config.KeyFile
	ws.server.LittleEndian = LittleEndian
	ws.server.NewAgent = newAgent

	processor = json.NewProcessor()

	if ws.server != nil {
		ws.server.Start()
	}

	if onInitCallback != nil {
		onInitCallback()
	}
}

func (ws *WsServerWrapper) Stop() {
	if ws.server != nil {
		ws.server.Close()
	}
}

// -------------------------------------------------------------------------------------
// Udp server

type UdpServerWrapper struct {
	server *network.UDPServer
}

func (udp *UdpServerWrapper) GetType() uint {
	return TYPE_SERVER_UDP
}

func (udp *UdpServerWrapper) GetAddr() string {
	return udp.server.Addr
}

func (udp *UdpServerWrapper) GetPool() *pool.ObjectPool {
	return nil
}

func (udp *UdpServerWrapper) Start() {
	config := conf.GetUDP()
	createUdpAgentPool()
	udp.server = new(network.UDPServer)
	udp.server.MaxConnNum = config.MaxConnNum
	udp.server.MinMsgLen = config.MinMsgLen
	udp.server.MaxMsgLen = config.MaxMsgLen
	udp.server.LittleEndian = LittleEndian
	udp.server.NewAgent = newUdpAgent
	udp.server.Start(config.Addr)
}

func (udp *UdpServerWrapper) Stop() {
	if udp.server != nil {
		udp.server.Close()
	}
	removeUdpAgentPool()
}
