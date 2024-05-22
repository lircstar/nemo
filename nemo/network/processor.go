package network

import (
	"reflect"
)

type MsgHandler func(Agent, []any)

type Processor interface {

	// Route must goroutine safe
	Route(agent Agent, msg any, userData any) error
	// Unmarshal must goroutine safe
	Unmarshal(data []byte) (any, error)
	// Marshal must goroutine safe
	Marshal(msg any) ([][]byte, error)

	// GetByteOrder get current message buffer's bytes order.
	GetByteOrder() bool

	// Register register message into processor.
	Register(msg any)

	// SetHandler set message handling function
	SetHandler(msg any, msgHandler MsgHandler)

	// SetRawHandler set raw message handling function.
	SetRawHandler(id uint16, msgRawHandler MsgHandler)

	// Range show all registered message.
	Range(f func(id uint16, t reflect.Type))

	// GetMsgId get current message id by type.
	GetMsgId(msgType reflect.Type) uint16
}
