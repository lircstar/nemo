package pool

import (
	"errors"
	"sync"
)

// ChanPool represents a pool of reusable channels
type ChanPool struct {
	mu    sync.Mutex
	pool  chan chan any
	size  int
	width int
}

// NewChanPool creates a new ChanPool
func NewChanPool(size, width int) (*ChanPool, error) {
	if size <= 0 || width <= 0 {
		return nil, errors.New("size and width must be positive")
	}

	pool := &ChanPool{
		pool:  make(chan chan any, size),
		size:  size,
		width: width,
	}

	for i := 0; i < size; i++ {
		pool.pool <- make(chan any, width)
	}

	return pool, nil
}

// Get retrieves a channel from the pool
func (p *ChanPool) Get() chan any {
	select {
	case ch := <-p.pool:
		return ch
	default:
		return make(chan any, p.width)
	}
}

// Put returns a channel to the pool
func (p *ChanPool) Put(ch chan any) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Clear the channel
	for len(ch) > 0 {
		<-ch
	}

	select {
	case p.pool <- ch:
	default:
		// Pool is full, discard the channel
	}
}

// Len returns the number of channels currently in the pool
func (p *ChanPool) Len() int {
	return len(p.pool)
}

// Close closes all channels in the pool
func (p *ChanPool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for {
		select {
		case ch := <-p.pool:
			close(ch)
		default:
			close(p.pool)
			return
		}
	}
}
