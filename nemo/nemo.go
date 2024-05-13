package nemo

import (
	"math"
	"nemo/sys/pool"
	"time"

	"nemo/nemo/conf"
	"nemo/nemo/network"
	"nemo/nemo/network/proto"
	"nemo/sys/log"
)

//-------------------------------------------------------------------------------------
// interface function.

func RegisterProcessor(cb ProcessCallBack) {
	registerProcessor = cb
}

func RegisterMessage(msg interface{}, msgHandler network.MsgHandler) {
	processor.Register(msg)
	processor.SetHandler(msg, msgHandler)

}

func RegisterRawMessage(id uint16, msgHandler network.MsgHandler) {
	processor.SetRawHandler(id, msgHandler)
}

func RegisterMessageNoHandler(msg interface{}) {
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

//-------------------------------------------------------------------------------------
// Define

type ProcessCallBack func() network.Processor

var registerProcessor func() network.Processor
var processor network.Processor

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
// Server

const (
	TYPE_SERVER_TCP      = 1
	TYPE_SEVER_WEBSOCKET = 2
)

var nemoType int

// Start tcp server.
func StartTcpServer() {
	Run(TYPE_SERVER_TCP)
}

// Start web socket server.
func StartWebSockServer() {
	Run(TYPE_SEVER_WEBSOCKET)
}

func Run(serverType int) {

	nemoType = serverType

	logInit()

	monitor()

	newTcpAgentPool()

	go mainProc()

	if nemoType == TYPE_SERVER_TCP {
		run()
	} else if nemoType == TYPE_SEVER_WEBSOCKET {
		webRun()
	}

	log.Infof("Nemo %v starting up.", version)
	if nemoType == TYPE_SERVER_TCP {
		log.Infof("Listen Address: %v", tcpServer.Addr)
	} else if nemoType == TYPE_SEVER_WEBSOCKET {
		log.Infof("Listen Address: %v", wsServer.Addr)
	}

	// close
	closeSig()

	logClose()

	// free agent pool
	delTcpAgentPool()
}

func destroy() {

	if onDestroyCallback != nil {
		onDestroyCallback()
	}

	if nemoType == TYPE_SERVER_TCP {
		if tcpServer != nil {
			tcpServer.Close()
		}
	} else if nemoType == TYPE_SEVER_WEBSOCKET {
		if wsServer != nil {
			wsServer.Close()
		}
	}
}

// -------------------------------------------------------------------------------------
// Tcp server.
var tcpServer *network.TCPServer

func run() {

	if len(conf.TCPAddr) == 0 {
		log.Error("ip adrress of server cannot be zero.")
		return
	}

	tcpServer = new(network.TCPServer)
	tcpServer.Addr = conf.TCPAddr
	tcpServer.MaxConnNum = conf.TcpMaxConnNum
	tcpServer.MaxConnNum = conf.TcpMinMsgLen
	tcpServer.MaxMsgLen = conf.TcpMaxMsgLen
	tcpServer.PendingWriteNum = 100
	tcpServer.NewAgent = newTcpAgent
	tcpServer.LenMsgLen = conf.LenMsgLen
	tcpServer.LittleEndian = conf.LittleEndian
	if registerProcessor == nil {
		processor = protobuf.NewProcessor()
	} else {
		processor = registerProcessor()
	}

	if tcpServer != nil {
		tcpServer.Start()
	}

	if onInitCallback != nil {
		onInitCallback()
	}

}

func newTcpAgent(conn *network.TCPConn) network.Agent {
	if tcpAgentPool.FreeCount() <= 1 {
		for i := 0; i < 128; i++ {
			tcpAgentPool.Create(new(TcpAgent))
		}
	}
	a := tcpAgentPool.Get().(*TcpAgent)
	a.conn = conn
	a.idleTime = time.Now().Unix()
	return a
}

// -------------------------------------------------------------------------------------
// Websocket server.
var wsServer *network.WSServer

func webRun() {
	if conf.WSAddr != "" {
		wsServer = new(network.WSServer)
		wsServer.Addr = conf.WSAddr
		wsServer.MaxConnNum = conf.TcpMaxConnNum
		wsServer.MaxMsgLen = conf.TcpMaxMsgLen
		wsServer.PendingWriteNum = conf.PendingWriteNum
		wsServer.HTTPTimeout = conf.HTTPTimeout
		wsServer.CertFile = conf.CertFile
		wsServer.KeyFile = conf.KeyFile
		wsServer.LittleEndian = conf.LittleEndian
		wsServer.NewAgent = newWebAgent
	}

	if registerProcessor == nil {
		processor = protobuf.NewProcessor()
	} else {
		processor = registerProcessor()
	}

	if wsServer != nil {
		wsServer.Start()
	}

	if onInitCallback != nil {
		onInitCallback()
	}
}

func newWebAgent(conn *network.WSConn) network.Agent {
	if tcpAgentPool.FreeCount() <= 1 {
		for i := 0; i < 128; i++ {
			tcpAgentPool.Create(new(TcpAgent))
		}
	}
	a := tcpAgentPool.Get().(*TcpAgent)
	a.conn = conn
	a.idleTime = time.Now().Unix()
	return a
}

//-------------------------------------------------------------------------------------
// Connection manager to connect other server.

// create a client and connect to a server.
func Connect(addr string) *network.TCPClient {
	client := new(network.TCPClient)
	client.Addr = addr
	client.AutoReconnect = conf.TcpReconnect
	client.ConnectInterval = 3 * time.Second
	client.PendingWriteNum = conf.PendingWriteNum
	client.LenMsgLen = conf.LenMsgLen
	client.MaxMsgLen = math.MaxInt32
	client.LittleEndian = conf.LittleEndian
	client.NewAgent = newTcpClientAgent
	// If have no processor create by server, create it by itself.
	if processor == nil {
		if registerProcessor == nil {
			processor = protobuf.NewProcessor()
		} else {
			processor = registerProcessor()
		}
	}
	client.Start()
	return client
}

// close all connection.
func Close(client *network.TCPClient) {
	if client != nil {
		client.Close()
	}
}

func newTcpClientAgent(conn *network.TCPConn) network.Agent {
	a := new(TcpAgent)
	a.conn = conn
	a.idleTime = time.Now().Unix()
	return a
}

// -------------------------------------------------------------------------------------
// Udp server
func CreateUDPServer(addr string) *network.UDPServer {
	if udpAgentPool == nil {
		udpAgentPool = pool.NewObjectPool()
	}
	server := new(network.UDPServer)
	server.MaxConnNum = conf.UdpMaxConnNum
	server.MinMsgLen = conf.UdpMinMsgLen
	server.MaxMsgLen = conf.UdpMaxMsgLen
	server.LittleEndian = conf.LittleEndian
	server.NewAgent = newUdpAgent
	server.Start(addr)
	return server
}

func newUdpAgent(conn network.Conn) network.Agent {
	if udpAgentPool.FreeCount() <= 1 {
		for i := 0; i < 128; i++ {
			udpAgentPool.Create(new(UdpAgent))
		}
	}
	a := udpAgentPool.Get().(*UdpAgent)
	a.conn = conn
	a.idleTime = time.Now().Unix()
	return a
}

// Udp client
func CreateUDPClient(addr string) {
	client := &network.UDPClient{}
	client.TimeOut = conf.UdpTimeout
	client.MinMsgLen = conf.UdpMinMsgLen
	client.MaxMsgLen = conf.UdpMaxMsgLen
	client.LittleEndian = conf.LittleEndian
	client.AutoReconnect = conf.UdpReconnect
	client.ConnectInterval = conf.UdpConnectInterval
	client.NewAgent = newUdpClientAgent
	client.Connect(addr)
}

func newUdpClientAgent(conn network.Conn) network.Agent {
	a := new(UdpAgent)
	a.conn = conn
	a.idleTime = time.Now().Unix()
	return a
}
