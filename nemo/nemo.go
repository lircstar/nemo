package nemo

import (
	"github.com/lircstar/nemo/nemo/network"
	"github.com/lircstar/nemo/nemo/server"
)

//-------------------------------------------------------------------------------------
// Server

func StartTCPServer() {
	server.New(new(server.TcpServerWrapper))
	server.Start()
}

func StartWSServer() {
	server.New(new(server.WsServerWrapper))
	server.Start()
}

func CreateUDPServer() server.Server {
	udp := server.New(new(server.UdpServerWrapper))
	udp.Start()
	return udp
}

//-------------------------------------------------------------------------------------
// Client

func Connect(addr string, style uint) network.Client {
	var ret network.Client
	if style == network.TYPE_CLIENT_TCP {
		client := new(server.TcpClientWrapper)
		ret = client.Connect(addr)
	} else if style == network.TYPE_CLIENT_WEBSOCKET {
		client := new(server.WsClientWrapper)
		ret = client.Connect(addr)
	} else if style == network.TYPE_CLIENT_UDP {
		client := new(server.UdpClientWrapper)
		ret = client.Connect(addr)
	}
	return ret
}
