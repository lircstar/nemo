package server

import (
	"github.com/lircstar/nemo/nemo/conf"
	"github.com/lircstar/nemo/nemo/network"
)

// -------------------------------------------------------------------------------------
// Define
const (
	TYPE_SERVER_TCP      = 1
	TYPE_SEVER_WEBSOCKET = 2
	TYPE_SERVER_UDP      = 3
)

type Server interface {
	GetType() uint
	GetAddr() string

	Start()
	Stop()
}

type Client interface {
	GetType() int

	Connect(addr string) *network.TCPClient
	Close()
}

type Event struct {
	agent    network.Agent
	msg      any
	userData any
}

var LittleEndian = conf.GetSYS().LittleEndian

var processor network.Processor

type ProcessCallBack func() network.Processor

var registerProcessor func() network.Processor

type EventCallback func()

// Server begin. (Initialize data before running server)
var onInitCallback EventCallback

// Server main loop
var onLoopCallback func()

// Server end. (For exiting server's withdraw)
var onDestroyCallback func()

type ConnectCallback func(network.Agent)

// Agent connected.
var onConnectCallback func(network.Agent)

// Agent closed.
var onCloseCallback func(network.Agent)

//-------------------------------------------------------------------------------------
// interface function.

func RegisterProcessor(pro network.Processor) {
	processor = pro
}

func RegisterMessage(msg any, msgHandler network.MsgHandler) {
	processor.Register(msg)
	processor.SetHandler(msg, msgHandler)

}

func RegisterRawMessage(id uint16, msgHandler network.MsgHandler) {
	processor.SetRawHandler(id, msgHandler)
}

func RegisterMessageNoHandler(msg any) {
	processor.Register(msg)
}

func RegisterRawMessageNoHandler(msg interface{}) {
	processor.Register(msg)
}

func RegisterOnInit(cb EventCallback) {
	onInitCallback = cb
}

func RegisterOnLoop(cb EventCallback) {
	onLoopCallback = cb
}

func RegisterOnDestroy(cb EventCallback) {
	onDestroyCallback = cb
}

func RegisterOnConnect(cb ConnectCallback) {
	onConnectCallback = cb
}

func RegisterOnClose(cb ConnectCallback) {
	onCloseCallback = cb
}
