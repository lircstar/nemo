package pool

import (
	"github.com/lircstar/nemo/sys/utest"
	"testing"
)

func Test_NewChannelPool(t *testing.T) {
	pool, err := NewChanPool(5, 10)
	utest.EqualNow(t, err, nil)
	utest.EqualNow(t, pool.Len(), 5)
}

func Test_ChannelPool_Get(t *testing.T) {
	pool, _ := NewChanPool(5, 10)
	ch := pool.Get()
	utest.EqualNow(t, cap(ch), 10)
	utest.EqualNow(t, pool.Len(), 4)
}

func Test_ChannelPool_Put(t *testing.T) {
	pool, _ := NewChanPool(5, 10)
	ch := pool.Get()
	pool.Put(ch)
	utest.EqualNow(t, pool.Len(), 5)
}

func Test_ChannelPool_Len(t *testing.T) {
	pool, _ := NewChanPool(5, 10)
	utest.EqualNow(t, pool.Len(), 5)
	pool.Get()
	utest.EqualNow(t, pool.Len(), 4)
}

func Test_ChannelPool_Close(t *testing.T) {
	pool, _ := NewChanPool(5, 10)
	pool.Close()
	utest.EqualNow(t, pool.Len(), 0)
}
