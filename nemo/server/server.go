package server

import (
	"github.com/lircstar/nemo/nemo/conf"
	"github.com/lircstar/nemo/nemo/network"
	"github.com/lircstar/nemo/nemo/network/json"
	protobuf "github.com/lircstar/nemo/nemo/network/proto"
	"github.com/lircstar/nemo/sys/log"
	"github.com/lircstar/nemo/sys/pool"
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

	if len(conf.TCPAddr) == 0 {
		log.Error("ip adrress of server cannot be zero.")
		return
	}

	tcp.server = new(network.TCPServer)
	tcp.server.Addr = conf.TCPAddr
	tcp.server.MaxConnNum = conf.TcpMaxConnNum
	tcp.server.MaxConnNum = conf.TcpMinMsgLen
	tcp.server.MaxMsgLen = conf.TcpMaxMsgLen
	tcp.server.PendingWriteNum = 100
	tcp.server.NewAgent = newAgent
	tcp.server.LenMsgLen = conf.LenMsgLen
	tcp.server.LittleEndian = conf.LittleEndian

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
	if len(conf.WSAddr) == 0 {
		log.Error("adrress of server cannot be zero.")
		return
	}

	ws.server = new(network.WSServer)
	ws.server.Addr = conf.WSAddr
	ws.server.MaxConnNum = conf.TcpMaxConnNum
	ws.server.MaxMsgLen = conf.TcpMaxMsgLen
	ws.server.PendingWriteNum = conf.PendingWriteNum
	ws.server.HTTPTimeout = conf.HTTPTimeout
	ws.server.CertFile = conf.CertFile
	ws.server.KeyFile = conf.KeyFile
	ws.server.LittleEndian = conf.LittleEndian
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
	createUdpAgentPool()
	udp.server = new(network.UDPServer)
	udp.server.MaxConnNum = conf.UdpMaxConnNum
	udp.server.MinMsgLen = conf.UdpMinMsgLen
	udp.server.MaxMsgLen = conf.UdpMaxMsgLen
	udp.server.LittleEndian = conf.LittleEndian
	udp.server.NewAgent = newUdpAgent
	udp.server.Start(conf.UDPAddr)
}

func (udp *UdpServerWrapper) Stop() {
	if udp.server != nil {
		udp.server.Close()
	}
	removeUdpAgentPool()
}
