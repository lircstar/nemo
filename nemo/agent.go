package nemo

import (
	"encoding/binary"
	"net"
	"reflect"
	"time"

	"nemo/nemo/conf"
	"nemo/nemo/network"
	"nemo/sys/log"
	"nemo/sys/pool"
)

// TcpAgent //////////////////////////////////////////////////////////
type TcpAgent struct {
	conn     network.Conn
	id       uint64
	idleTime int64
	//outFlag  bool // it is a flag of connection connect to other server.
	userData interface{}
}

type Event struct {
	agent    network.Agent
	msg      interface{}
	userData interface{}
}

//func (a *Agent) IsOutFlag() bool {
//	return a.outFlag
//}

func routeMessage(agent network.Agent, msg interface{}, userData interface{}) error {
	return processor.Route(agent, msg, userData)
}

// Run goroutine safe
func (a *TcpAgent) Run(data []byte) {
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
			if conf.TcpRoutineSafe {
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

func (a *TcpAgent) SendMessage(msg interface{}) bool {
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

func (a *TcpAgent) SendRawMessage(id uint16, msg []byte) bool {

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
func (a *TcpAgent) OnConnect() {
	if onConnectCallback != nil {
		onConnectCallback(a)
	}
}

// OnClose goroutine safe
func (a *TcpAgent) OnClose() {
	if onCloseCallback != nil {
		onCloseCallback(a)
	}
	// free agent from pool.
	if tcpAgentPool != nil {
		tcpAgentPool.Free(a)
	}
}

func (a *TcpAgent) Close() {
	a.conn.Close()
}

func (a *TcpAgent) Destroy() {
	a.conn.Destroy()
}

func (a *TcpAgent) LocalAddr() net.Addr {
	return a.conn.LocalAddr()
}

func (a *TcpAgent) RemoteAddr() net.Addr {
	return a.conn.RemoteAddr()
}

func (a *TcpAgent) UserData() interface{} {
	return a.userData
}

func (a *TcpAgent) SetUserData(data interface{}) {
	a.userData = data
}

func (a *TcpAgent) SetConnectionId(id uint64) {
	a.id = id
}

func (a *TcpAgent) ConnectionId() uint64 {
	return a.id
}

//////////////////////////////////////////////////////////////////
// TcpAgent Pool

var tcpAgentPool *pool.ObjectPool = nil

func newTcpAgentPool() *pool.ObjectPool {
	tcpAgentPool = pool.NewObjectPool()
	return tcpAgentPool
}

func delTcpAgentPool() {
	tcpAgentPool.Range(func(i interface{}) {
		if i != nil {
			i = nil
		}
	})
	tcpAgentPool = nil
}

//////////////////////////////////////////////////////////////////
// UdpAgent

type UdpAgent struct {
	conn     network.Conn
	id       uint64
	idleTime int64
	//outFlag  bool // it is a flag of connection connect to other server.
	userData interface{}
}

// Run goroutine safe
func (a *UdpAgent) Run(data []byte) {
	if processor != nil {
		msg, err := processor.Unmarshal(data)
		if err != nil {
			log.Warnf("unmarshal message error: %v", err)
			return
		}
		// main loop
		if conf.UdpRoutineSafe {
			eventChan <- &Event{a, msg, a.userData}
		} else {
			err = processor.Route(a, msg, a.userData)
		}

		if err != nil {
			log.Warnf("route message error: %v", err)
			return
		}
	}

	a.idleTime = time.Now().Unix()

}

func (a *UdpAgent) SendMessage(msg interface{}) bool {
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

func (a *UdpAgent) SendRawMessage(id uint16, msg []byte) bool {

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

// goroutine safe
func (a *UdpAgent) OnConnect() {
	if onConnectCallback != nil {
		onConnectCallback(a)
	}
}

// goroutine safe
func (a *UdpAgent) OnClose() {
	if onCloseCallback != nil {
		onCloseCallback(a)
	}
	// free agent from pool.
	if udpAgentPool != nil {
		udpAgentPool.Free(a)
	}

}

func (a *UdpAgent) Close() {
	a.conn.Close()
}

func (a *UdpAgent) Destroy() {
	a.conn.Destroy()
}

func (a *UdpAgent) LocalAddr() net.Addr {
	return a.conn.LocalAddr()
}

func (a *UdpAgent) RemoteAddr() net.Addr {
	return a.conn.RemoteAddr()
}

func (a *UdpAgent) UserData() interface{} {
	return a.userData
}

func (a *UdpAgent) SetUserData(data interface{}) {
	a.userData = data
}

func (a *UdpAgent) SetConnectionId(id uint64) {
	a.id = id
}

func (a *UdpAgent) ConnectionId() uint64 {
	return a.id
}

//////////////////////////////////////////////////////////////////
// TcpAgent Pool

var udpAgentPool *pool.ObjectPool = nil

func newUdpAgentPool() *pool.ObjectPool {
	udpAgentPool = pool.NewObjectPool()
	return udpAgentPool
}

func delUdpAgentPool() {
	udpAgentPool.Range(func(i interface{}) {
		if i != nil {
			i = nil
		}
	})
	udpAgentPool = nil
}
