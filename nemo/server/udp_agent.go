package server

import (
	"github.com/lircstar/nemo/nemo/conf"
	"github.com/lircstar/nemo/sys/log"
	"time"
)

//////////////////////////////////////////////////////////////////
// UdpAgent

type UdpAgent struct {
	Agent
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
		if conf.GetUDP().RoutineSafe {
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

func (a *UdpAgent) IsLive() bool {
	if int(time.Now().Unix()-a.idleTime) > conf.GetUDP().TimeOut {
		return false
	} else if a.conn == nil || !a.conn.IsClosed() {
		return false
	}
	return true
}

func (a *UdpAgent) OnClose() {
	if onCloseCallback != nil {
		onCloseCallback(a)
	}
	// free agent from pool.
	delUdpAgent(a)
}
