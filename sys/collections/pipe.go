package collections

import (
	"sync"
)

// 不限制大小，添加不发生阻塞，接收阻塞等待
type Pipe struct {
	list      []any
	listGuard sync.Mutex
	listCond  *sync.Cond
}

// 添加时不会发送阻塞
func (pipe *Pipe) Add(msg any) {
	pipe.listGuard.Lock()
	pipe.list = append(pipe.list, msg)
	pipe.listGuard.Unlock()

	pipe.listCond.Signal()
}

func (pipe *Pipe) Reset() {
	pipe.list = pipe.list[0:0]
}

// 如果没有数据，发生阻塞
func (pipe *Pipe) Pick(retList *[]any) (exit bool) {

	pipe.listGuard.Lock()

	for len(pipe.list) == 0 {
		pipe.listCond.Wait()
	}

	pipe.listGuard.Unlock()

	pipe.listGuard.Lock()

	// 复制出队列

	for _, data := range pipe.list {

		if data == nil {
			exit = true
			break
		} else {
			*retList = append(*retList, data)
		}
	}

	pipe.Reset()
	pipe.listGuard.Unlock()

	return
}

func NewPipe() *Pipe {
	pipe := &Pipe{}
	pipe.listCond = sync.NewCond(&pipe.listGuard)

	return pipe
}
