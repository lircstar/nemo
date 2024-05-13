package network

import "net"

type Agent interface {
	OnConnect()
	Run(data []byte)
	OnClose()

	SendMessage(msg interface{}) bool
	SendRawMessage(id uint16, msg []byte) bool

	LocalAddr() net.Addr
	RemoteAddr() net.Addr

	Close()
	Destroy()

	SetConnectionId(id uint64)
	ConnectionId() uint64

	SetUserData(data interface{})
	UserData() interface{}
}
