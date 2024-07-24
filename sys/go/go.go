package g

import (
	"container/list"
	"sync"

	"nemo/sys/log"
)

// one Go per goroutine (goroutine not safe)
type Go struct {
	ChanCb    chan func()
	pendingGo int
}

type LinearGo struct {
	f  func()
	cb func()
}

type LinearContext struct {
	g              *Go
	linearGo       *list.List
	mutexLinearGo  sync.Mutex
	mutexExecution sync.Mutex
}

func New(l int) *Go {
	g := new(Go)
	g.ChanCb = make(chan func(), l)
	return g
}

func (g *Go) Go(f func(), cb func()) {
	g.pendingGo++

	go func() {
		defer func() {
			g.ChanCb <- cb

		}()

		f()
	}()
}

func (g *Go) Cb() {
	defer func() {
		g.pendingGo--
		if r := recover(); r != nil {
			log.Recover()
		}
	}()

	cb := <-g.ChanCb
	if cb != nil {
		cb()
	}
}

func (g *Go) Close() {
	for g.pendingGo > 0 {
		g.Cb()
	}
}

func (g *Go) Idle() bool {
	return g.pendingGo == 0
}

func (g *Go) NewLinearContext() *LinearContext {
	c := new(LinearContext)
	c.g = g
	c.linearGo = list.New()
	return c
}

func (c *LinearContext) Go(f func(), cb func()) {
	c.g.pendingGo++

	c.mutexLinearGo.Lock()
	c.linearGo.PushBack(&LinearGo{f: f, cb: cb})
	c.mutexLinearGo.Unlock()

	go func() {
		c.mutexExecution.Lock()
		defer c.mutexExecution.Unlock()

		c.mutexLinearGo.Lock()
		e := c.linearGo.Remove(c.linearGo.Front()).(*LinearGo)
		c.mutexLinearGo.Unlock()

		defer func() {
			c.g.ChanCb <- e.cb
			log.Recover()
		}()

		e.f()
	}()
}
