package server

import (
	"github.com/lircstar/nemo/nemo/conf"
	"github.com/lircstar/nemo/nemo/network"
	"os"
	"os/signal"
	"runtime/pprof"
	"strconv"
	"time"

	"github.com/lircstar/nemo/sys/log"
)

var server Server = nil

var eventChan = make(chan *Event, 1024)
var exitProcChan = make(chan int, 1)
var endProcChan = make(chan int, 1)

func New(s Server) Server {
	server = s
	return server
}

func mainProc() {
	t1 := time.NewTimer(time.Second * 10)
	for {
		select {
		case event := <-eventChan:
			err := processor.Route(event.agent, event.msg, event.userData)
			if err != nil {
				log.Debugf("route message error: %v", err)
			}
		case <-time.After(time.Millisecond * 30):
			onLoopCallback()
		case <-t1.C:
			if server.GetType() == TYPE_SERVER_TCP {
				loopAgentPool()
				loopUdpAgentPool()
			}
			t1.Reset(time.Second * 10)
		case <-exitProcChan:
			doFinish()
			return
		}

	}
}

func doFinish() {
	destroy()
	endProcChan <- 0
	log.Info("Nemo closed.")
}

func loopAgentPool() {
	agentPool.UsedRange(func(i any) {
		agent := i.(network.Agent)
		if agent != nil {
			conn := agent.GetConn()
			if conn != nil && !conn.IsClosed() {
				timeout := conf.GetTCP().TimeOut
				if timeout > 0 && !agent.IsActive() &&
					time.Now().Unix()-agent.GetIdleTime() > int64(timeout) {
					agent.Close()
				}
			}
		}
	})
}

func loopUdpAgentPool() {
	if udpAgentPool != nil {
		udpAgentPool.UsedRange(func(i interface{}) {
			agent := i.(*UdpAgent)
			if agent != nil {
				if agent.conn != nil && !agent.conn.IsClosed() {
					timeout := conf.GetUDP().TimeOut
					if timeout > 0 &&
						time.Now().Unix()-agent.idleTime > int64(timeout) {
						agent.Close()
					}
				}
			}
		})
	}
}

//////////////////////////////////////////////////////////////
// nemo init utility.

func logInit() {
	config := conf.GetSYS()
	if config.LogLevel != "" {
		log.SetLevel(config.LogLevel)
	}
	if config.LogFile {
		log.LogFile()
	}
}

func logClose() {
	log.Close()
}

func monitor() {
	config := conf.GetSYS()
	monitor, err := strconv.ParseInt(config.Monitor, 2, 32)
	if err != nil {
		log.Warnf("Failed to monitor; %v", err)
	} else {
		if monitor&1 == 1 {
			cpu, _ := os.Create("cpu.pprof")
			defer cpu.Close()
			err := pprof.StartCPUProfile(cpu)
			if err != nil {
				log.Fatal(err)
			}
			defer pprof.StopCPUProfile()
		}

		if monitor&2 == 2 {
			mem, err := os.Create("mem.pprof")
			if err != nil {
				log.Fatal(err)
			}
			if err = pprof.WriteHeapProfile(mem); err != nil {
				log.Warnf("write memory profile %v", err)
			}
			defer mem.Close()
		}

		if monitor&4 == 4 {
			block, err := os.Create("block.pprof")
			if err != nil {
				log.Fatal(err)
			}
			if err = pprof.Lookup("block").WriteTo(block, 0); err != nil {
				log.Warnf("write block profile %v", err)
			}
			defer block.Close()
		}

		if monitor&8 == 8 {
			gr, err := os.Create("go.pprof")
			if err != nil {
				log.Fatal(err)
			}
			if err = pprof.Lookup("goroutine").WriteTo(gr, 0); err != nil {
				log.Warnf("write goroutine profile %v", err)
			}
			defer gr.Close()
		}
	}
}

func closeSig() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	sig := <-c
	log.Infof("Nemo closing. (signal:%v)", sig)

	exitProcChan <- 0
	<-endProcChan
}

// ////////////////////////////////////////////////////////////
// server

func Start() {

	logInit()

	monitor()

	createAgentPool()

	go mainProc()

	server.Start()

	log.Infof("Nemo %v starting up.", conf.GetSYS().Version)
	log.Infof("Listen Address: %v", server.GetAddr())

	// close
	closeSig()

	removeAgentPool()

	logClose()

}

func Stop() {
	exitProcChan <- 0
}

func destroy() {

	if onDestroyCallback != nil {
		onDestroyCallback()
	}

	server.Stop()
}
