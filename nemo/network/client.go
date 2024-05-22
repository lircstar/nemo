package network

const (
	TYPE_CLIENT_TCP       = 1001
	TYPE_CLIENT_WEBSOCKET = 1002
	TYPE_CLIENT_UDP       = 1003
)

type Client interface {
	Start()
	Send(msg any) bool
	Close()

	GetType() uint
	GetConnected() bool
	GetAgent() Agent
	GetAddress() string
}
