package network

type MsgHandler func(Agent, []any)

type Processor interface {

	// Route must goroutine safe
	Route(agent Agent, msg any, userData any) error
	// Unmarshal must goroutine safe
	Unmarshal(data []byte) (any, error)
	// Marshal must goroutine safe
	Marshal(msg any) ([][]byte, error)

	// Register register message into processor.
	Register(msg any)

	// SetHandler set message handling function
	SetHandler(msg any, msgHandler MsgHandler)

	// SetRawHandler set raw message handling function.
	SetRawHandler(id uint16, msgRawHandler MsgHandler)

	// Range show all registered message.
	Range(f func(id uint16, name string))

	// GetMsgId get current message id by type.
	GetMsgId(msgType any) uint16
}
