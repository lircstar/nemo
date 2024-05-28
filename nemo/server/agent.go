package server

import (
	"encoding/binary"
	"github.com/lircstar/nemo/nemo/conf"
	"github.com/lircstar/nemo/nemo/network"
	"github.com/lircstar/nemo/sys/log"
	"github.com/lircstar/nemo/sys/pool"
	"net"
	"reflect"
	"time"
)

type Agent struct {
	style    uint
	conn     network.Conn
	id       uint64
	idleTime int64
	pool     *pool.ObjectPool
	//outFlag  bool // it is a flag of connection connect to other server.
	userData any
}

func (a *Agent) GetType() uint {
	return a.style
}

func (a *Agent) SetType(style uint) {
	a.style = style
}

func (a *Agent) GetConn() network.Conn {
	return a.conn
}

func (a *Agent) GetIdleTime() int64 {
	return a.idleTime
}

// Run goroutine safe
func (a *Agent) Run(data []byte) {
	for {
		data, err := a.conn.ReadMsg()
		if err != nil {
			log.Debugf("read message: %v", err)
			break
		}

		if processor != nil {
			msg, err := processor.Unmarshal(data)
			if err != nil {
				log.Warnf("unmarshal message error: %v", err)
				break
			}
			// main loop
			if conf.RoutineSafe {
				eventChan <- &Event{a, msg, a.userData}
			} else {
				err = processor.Route(a, msg, a.userData)
			}

			if err != nil {
				log.Warnf("route message error: %v", err)
				break
			}
		}
		a.idleTime = time.Now().Unix()
	}
}

func (a *Agent) SendMessage(msg any) bool {
	if processor != nil {
		data, err := processor.Marshal(msg)
		if err != nil {
			log.Errorf("marshal message %v error: %v", reflect.TypeOf(msg), err)
			return false
		}
		err = a.conn.WriteMsg(data...)
		if err != nil {
			log.Errorf("write message %v error: %v", reflect.TypeOf(msg), err)
			return false
		}
		return true
	}
	return false
}

func (a *Agent) SendRawMessage(id uint16, msg []byte) bool {

	_id := make([]byte, 2)
	if processor.GetByteOrder() {
		binary.LittleEndian.PutUint16(_id, id)
	} else {
		binary.BigEndian.PutUint16(_id, id)
	}

	err := a.conn.WriteMsg(_id, msg)
	if err != nil {
		log.Errorf("write message %v error: %v", id, err)
		return false
	}
	return true
}

// OnConnect goroutine safe
func (a *Agent) OnConnect() {
	if onConnectCallback != nil {
		onConnectCallback(a)
	}
}

// OnClose goroutine safe
func (a *Agent) OnClose() {
	if onCloseCallback != nil {
		onCloseCallback(a)
	}
	// free agent from pool.
	delAgent(a)
}

func (a *Agent) Close() {
	a.conn.Close()
}

func (a *Agent) Destroy() {
	a.conn.Destroy()
}

func (a *Agent) LocalAddr() net.Addr {
	return a.conn.LocalAddr()
}

func (a *Agent) RemoteAddr() net.Addr {
	return a.conn.RemoteAddr()
}

func (a *Agent) UserData() any {
	return a.userData
}

func (a *Agent) SetUserData(data any) {
	a.userData = data
}

func (a *Agent) SetConnectionId(id uint64) {
	a.id = id
}

func (a *Agent) ConnectionId() uint64 {
	return a.id
}
