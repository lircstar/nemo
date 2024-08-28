package network

import "net"

const (
	TYPE_AGENT_TCP       = 1
	TYPE_AGENT_WEBSOCKET = 2
	TYPE_AGENT_UDP       = 3
)

type Agent interface {
	GetType() uint
	SetType(style uint)

	IsActive() bool

	GetConn() Conn
	GetIdleTime() int64

	OnConnect()
	Run(data []byte)
	OnClose()

	SendMessage(msg any) bool
	SendRawMessage(id uint16, msg []byte) bool

	LocalAddr() net.Addr
	RemoteAddr() net.Addr

	Close()
	Destroy()

	SetConnectionId(id uint64)
	ConnectionId() uint64

	SetUserData(data any)
	UserData() any
}
