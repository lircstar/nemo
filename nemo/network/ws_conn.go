package network

import (
	"errors"
	"github.com/gorilla/websocket"
	"nemo/sys/log"
	"net"
	"sync/atomic"
)

type WebsocketConnSet map[*websocket.Conn]struct{}

type WSConn struct {
	//ConnOption

	conn      *websocket.Conn
	writeChan chan []byte
	maxMsgLen int
	closeFlag atomic.Bool
}

func newWSConn(conn *websocket.Conn, pendingWriteNum int, maxMsgLen int) *WSConn {
	wsConn := new(WSConn)
	//wsConn.conn = conn
	wsConn.writeChan = make(chan []byte, pendingWriteNum)
	wsConn.maxMsgLen = maxMsgLen
	return wsConn
}

func (wsConn *WSConn) start() {
	go func() {
		for b := range wsConn.writeChan {
			if b == nil {
				break
			}

			err := wsConn.conn.WriteMessage(websocket.TextMessage, b)
			if err != nil {
				break
			}
		}

		_ = wsConn.conn.Close()
		wsConn.closeFlag.Store(true)
	}()
}

func (wsConn *WSConn) doDestroy() {
	wsConn.conn.UnderlyingConn().(*net.TCPConn).SetLinger(0)
	wsConn.conn.Close()

	if !wsConn.closeFlag.Load() {
		close(wsConn.writeChan)
		wsConn.closeFlag.Store(true)
	}
}

func (wsConn *WSConn) Destroy() {
	wsConn.doDestroy()
}

func (wsConn *WSConn) Close() {

	if wsConn.closeFlag.Load() {
		return
	}

	wsConn.doWrite(nil)
	wsConn.closeFlag.Store(true)
}

func (wsConn *WSConn) doWrite(b []byte) {
	if len(wsConn.writeChan) == cap(wsConn.writeChan) {
		log.Info("close conn: channel full")
		wsConn.doDestroy()
		return
	}

	wsConn.writeChan <- b
}

func (wsConn *WSConn) LocalAddr() net.Addr {
	return wsConn.conn.LocalAddr()
}

func (wsConn *WSConn) RemoteAddr() net.Addr {
	return wsConn.conn.RemoteAddr()
}

func (WSConn *WSConn) IsClosed() bool {
	return WSConn.closeFlag.Load()
}

// goroutine not safe
func (wsConn *WSConn) ReadMsg() ([]byte, error) {
	//fmt.Printf("read msg %v \n", wsConn.conn)
	_, b, err := wsConn.conn.ReadMessage()
	return b, err
}

// args must not be modified by the others goroutines
func (wsConn *WSConn) WriteMsg(args ...[]byte) error {
	if wsConn.closeFlag.Load() {
		return nil
	}

	// get len
	var msgLen int
	for i := 0; i < len(args); i++ {
		msgLen += len(args[i])
	}

	// check len
	if msgLen > wsConn.maxMsgLen {
		return errors.New("message too long")
	} else if msgLen < 1 {
		return errors.New("message too short")
	}

	// don't copy
	if len(args) == 1 {
		wsConn.doWrite(args[0])
		return nil
	}

	// merge the args
	msg := make([]byte, msgLen)
	l := 0
	for i := 0; i < len(args); i++ {
		copy(msg[l:], args[i])
		l += len(args[i])
	}

	wsConn.doWrite(msg)

	return nil
}
