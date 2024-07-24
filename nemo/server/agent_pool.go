package server

import (
	"nemo/nemo/network"
	"nemo/sys/pool"
	"time"
)

var agentPool *pool.ObjectPool

func createAgentPool() {
	if agentPool != nil {
		return
	}
	agentPool = pool.NewObjectPool()
}

func newAgent(conn network.Conn) network.Agent {
	if agentPool.FreeCount() <= 1 {
		for i := 0; i < 128; i++ {
			agentPool.Create(new(Agent))
		}
	}
	a := agentPool.Get().(*Agent)
	a.conn = conn
	a.idleTime = time.Now().Unix()
	return a
}

func delAgent(a network.Agent) {
	if agentPool != nil {
		agentPool.Free(a)
	}
}

func removeAgentPool() {
	if agentPool == nil {
		return
	}
	agentPool.Range(func(i any) {
		if i != nil {
			i = nil
		}
	})
	agentPool = nil
}

////////////////////////////////////////////////////////////////////
// UdpAgent Pool

var udpAgentPool *pool.ObjectPool = nil

func createUdpAgentPool() *pool.ObjectPool {
	if udpAgentPool != nil {
		return udpAgentPool
	}
	udpAgentPool = pool.NewObjectPool()
	return udpAgentPool
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

func delUdpAgent(a network.Agent) {
	if udpAgentPool != nil {
		udpAgentPool.Free(a)
	}
}

func removeUdpAgentPool() {
	if udpAgentPool == nil {
		return
	}
	udpAgentPool.Range(func(i any) {
		if i != nil {
			i = nil
		}
	})
	udpAgentPool = nil
}
