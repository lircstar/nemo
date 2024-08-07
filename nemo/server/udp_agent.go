package server

import (
	"nemo/nemo/conf"
	"nemo/sys/log"
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

func (a *UdpAgent) OnClose() {
	if onCloseCallback != nil {
		onCloseCallback(a)
	}
	// free agent from pool.
	delUdpAgent(a)
}
