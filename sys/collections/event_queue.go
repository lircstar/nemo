package collections

import (
	"log"
	"runtime/debug"
	"sync"
)

// 事件队列
type EventQueue interface {
	// 事件队列开始工作
	StartLoop() EventQueue

	// 停止事件队列
	StopLoop() EventQueue

	// 等待退出
	Wait()

	// 投递事件, 通过队列到达消费者端
	Post(callback func())

	// 是否捕获异常
	EnableCapturePanic(v bool)
}

type eventQueue struct {
	*Pipe

	endSignal sync.WaitGroup

	capturePanic bool
}

// 启动崩溃捕获
func (queue *eventQueue) EnableCapturePanic(v bool) {
	queue.capturePanic = v
}

// 派发事件处理回调到队列中
func (queue *eventQueue) Post(callback func()) {

	if callback == nil {
		return
	}

	queue.Add(callback)
}

// 保护调用用户函数
func (queue *eventQueue) protectedCall(callback func()) {

	if queue.capturePanic {
		defer func() {

			if err := recover(); err != nil {

				debug.PrintStack()
			}

		}()
	}

	callback()
}

// 开启事件循环
func (queue *eventQueue) StartLoop() EventQueue {

	queue.endSignal.Add(1)

	go func() {

		var writeList []any

		for {
			writeList = writeList[0:0]
			exit := queue.Pick(&writeList)

			// 遍历要发送的数据
			for _, msg := range writeList {
				switch t := msg.(type) {
				case func():
					queue.protectedCall(t)
				case nil:
					break
				default:
					log.Printf("unexpected type %T", t)
				}
			}

			if exit {
				break
			}
		}

		queue.endSignal.Done()
	}()

	return queue
}

// 停止事件循环
func (queue *eventQueue) StopLoop() EventQueue {
	queue.Add(nil)
	return queue
}

// 等待退出消息
func (queue *eventQueue) Wait() {
	queue.endSignal.Wait()
}

// 创建默认长度的队列
func NewEventQueue() EventQueue {

	return &eventQueue{
		Pipe: NewPipe(),
	}
}

//
//// 在会话对应的Peer上的事件队列中执行callback，如果没有队列，则马上执行
//func SessionQueuedCall(ses Session, callback func()) {
//	if ses == nil {
//		return
//	}
//	q := ses.Peer().(interface {
//		Queue() EventQueue
//	}).Queue()
//
//	QueuedCall(q, callback)
//}

// 有队列时队列调用，无队列时直接调用
func QueuedCall(queue EventQueue, callback func()) {
	if queue == nil {
		callback()
	} else {
		queue.Post(callback)
	}
}
