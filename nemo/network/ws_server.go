package network

import (
	"crypto/tls"
	"github.com/gorilla/websocket"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/lircstar/nemo/nemo/conf"
	"github.com/lircstar/nemo/sys/log"
	"github.com/lircstar/nemo/sys/pool"
)

type WSServer struct {
	Addr            string
	MaxConnNum      int
	PendingWriteNum int
	MaxMsgLen       int
	HTTPTimeout     time.Duration
	CertFile        string
	KeyFile         string
	LittleEndian    bool
	NewAgent        func(Conn) Agent
	ln              net.Listener
	handler         *WSHandler
}

type WSHandler struct {
	maxConnNum      int
	pendingWriteNum int
	maxMsgLen       int
	newAgent        func(Conn) Agent
	upgrader        websocket.Upgrader
	connPool        *pool.ObjectPool
	wg              sync.WaitGroup
}

func (handler *WSHandler) newWSConn(conn *websocket.Conn, pendingWriteNum int, maxMsgLen int) *WSConn {
	if handler.connPool.FreeCount() <= 1 {
		for i := 0; i < 128; i++ {
			wsConn := newWSConn(conn, pendingWriteNum, maxMsgLen)
			handler.connPool.Create(wsConn)
		}
	}

	wsConn := handler.connPool.Get().(*WSConn)
	wsConn.closeFlag.Store(false)
	wsConn.conn = conn
	return wsConn
}

func (handler *WSHandler) delWSConn(wsConn *WSConn) {
	if wsConn != nil {
		wsConn.Close()
		handler.connPool.Free(wsConn)
	}
}

func (handler *WSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	conn, err := handler.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Debugf("upgrade error: %v", err)
		return
	}
	conn.SetReadLimit(int64(handler.maxMsgLen))

	handler.wg.Add(1)
	defer handler.wg.Done()

	if handler.connPool == nil {
		conn.Close()
		return
	}
	if handler.connPool.UsedCount() >= handler.maxConnNum {
		conn.Close()
		log.Warn("too many connections")
		return
	}

	wsConn := handler.newWSConn(conn, handler.pendingWriteNum, handler.maxMsgLen)
	wsConn.start()
	agent := handler.newAgent(wsConn)
	agent.SetType(TYPE_AGENT_WEBSOCKET)
	agent.Run(nil)

	handler.delWSConn(wsConn)

	agent.OnClose()
}

func (server *WSServer) Start() {
	if server.Addr == "" {
		server.Addr = conf.WSAddr
	}

	ln, err := net.Listen("tcp", server.Addr)
	if err != nil {
		log.Fatalf("%v", err)
	}

	if server.MaxConnNum <= 0 {
		server.MaxConnNum = 100
		log.Warnf("invalid MaxConnNum, reset to %v", server.MaxConnNum)
	}
	server.PendingWriteNum = conf.PendingWriteNum
	if server.PendingWriteNum <= 0 {
		server.PendingWriteNum = 100
		log.Warnf("invalid PendingWriteNum, reset to %v", server.PendingWriteNum)
	}

	if server.MaxMsgLen <= 0 {
		server.MaxMsgLen = 4096
		log.Warnf("invalid MaxMsgLen, reset to %v", server.MaxMsgLen)
	}
	server.HTTPTimeout = conf.HTTPTimeout
	if server.HTTPTimeout <= 0 {
		server.HTTPTimeout = 10 * time.Second
		log.Warnf("invalid HTTPTimeout, reset to %v", server.HTTPTimeout)
	}
	if server.NewAgent == nil {
		log.Fatal("NewAgent must not be nil")
	}

	if server.CertFile != "" || server.KeyFile != "" {
		config := &tls.Config{}
		config.NextProtos = []string{"http/1.1"}

		var err error
		config.Certificates = make([]tls.Certificate, 1)
		config.Certificates[0], err = tls.LoadX509KeyPair(server.CertFile, server.KeyFile)
		if err != nil {
			log.Fatal("%v", err)
		}

		ln = tls.NewListener(ln, config)
	}

	server.ln = ln
	server.handler = &WSHandler{
		maxConnNum:      server.MaxConnNum,
		pendingWriteNum: server.PendingWriteNum,
		maxMsgLen:       server.MaxMsgLen,
		newAgent:        server.NewAgent,
		connPool:        pool.NewObjectPool(),
		upgrader: websocket.Upgrader{
			HandshakeTimeout: server.HTTPTimeout,
			CheckOrigin:      func(_ *http.Request) bool { return true },
		},
	}

	httpServer := &http.Server{
		Addr:           server.Addr,
		Handler:        server.handler,
		ReadTimeout:    server.HTTPTimeout,
		WriteTimeout:   server.HTTPTimeout,
		MaxHeaderBytes: 1024,
	}

	go httpServer.Serve(ln)
}

func (server *WSServer) Close() {
	server.ln.Close()

	server.handler.connPool.Range(func(i any) {
		if i != nil {
			conn := i.(*WSConn).conn
			if conn != nil {
				conn.Close()
			}
		}
	})

	server.handler.wg.Wait()

	server.handler.connPool = nil
}
