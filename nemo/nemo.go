package nemo

import (
	"nemo/nemo/network"
	"nemo/nemo/server"
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

func Connect(addr string, style uint) server.Client {
	var client server.Client
	if style == network.TYPE_CLIENT_TCP {
		client := new(server.TcpClientWrapper)
		client.Connect(addr)
	} else if style == network.TYPE_CLIENT_WEBSOCKET {
		client := new(server.WsClientWrapper)
		client.Connect(addr)
	} else if style == network.TYPE_CLIENT_UDP {
		client := new(server.UdpClientWrapper)
		client.Connect(addr)
	}
	return client
}
