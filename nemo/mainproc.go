package nemo

import (
	"nemo/nemo/conf"
	"os"
	"os/signal"
	"runtime/pprof"
	"strconv"
	"time"

	"nemo/sys/log"
)

var eventChan = make(chan *Event, 1024)
var exitProcChan = make(chan int, 1)
var endProcChan = make(chan int, 1)

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
			if nemoType == TYPE_SERVER_TCP {
				loopTcpAgentPool()
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

func loopTcpAgentPool() {
	tcpAgentPool.UsedRange(func(i interface{}) {
		agent := i.(*TcpAgent)
		if agent != nil {
			if agent.conn != nil && !agent.conn.IsClosed() {
				if conf.TcpTimeout > 0 &&
					time.Now().Unix()-agent.idleTime > int64(conf.TcpTimeout) {
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
					if conf.UdpTimeout > 0 &&
						time.Now().Unix()-agent.idleTime > int64(conf.UdpTimeout) {
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
	if conf.LogLevel != "" {
		log.SetLevel(conf.LogLevel)
	}
	if conf.LogFile {
		log.LogFile()
	}
}

func logClose() {
	log.Close()
}

func monitor() {
	monitor, err := strconv.ParseInt(conf.Monitor, 2, 32)
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
