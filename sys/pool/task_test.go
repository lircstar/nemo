package pool

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/lircstar/nemo/sys/utest"
)

type MockTask struct {
	runCalled    int32
	finishCalled int32
}

func (m *MockTask) Run() {
	atomic.AddInt32(&m.runCalled, 1)
}

func (m *MockTask) Finish() {
	atomic.AddInt32(&m.finishCalled, 1)
}

func Test_TaskPool_PushTask(t *testing.T) {
	pool := NewTaskPool()
	task := &MockTask{}
	pool.PushTask(task)

	time.Sleep(100 * time.Millisecond) // Allow some time for the goroutine to run

	utest.EqualNow(t, atomic.LoadInt32(&task.runCalled), int32(1))
}

func Test_TaskPool_getFinishedTask(t *testing.T) {
	pool := NewTaskPool()
	task := &MockTask{}
	pool.PushTask(task)

	time.Sleep(100 * time.Millisecond) // Allow some time for the goroutine to run

	finishedTask := pool.getFinishedTask()
	utest.NotNilNow(t, finishedTask)
	utest.EqualNow(t, finishedTask, task)
}

func Test_TaskPool_Loop(t *testing.T) {
	pool := NewTaskPool()
	task := &MockTask{}
	pool.PushTask(task)

	time.Sleep(100 * time.Millisecond) // Allow some time for the goroutine to run

	pool.Loop()
	utest.EqualNow(t, atomic.LoadInt32(&task.finishCalled), int32(1))
}
