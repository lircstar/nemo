package pool

import (
	"sync"
	"sync/atomic"
)

type ITask interface {
	Run()
	Finish()
}

type Task struct {
	pool       *TaskPool
	task       ITask
	activeChan chan ITask
}

func (t *Task) run() {
	task := <-t.activeChan
	task.Run()
	t.pool.pushFinishedTask(task)
	t.pool.pendingTaskCount.Add(-1)
}

func (t *Task) attachTask(task ITask) {
	t.task = task
	t.activeChan <- task
}

type TaskPool struct {
	pendingTaskCount atomic.Int32
	finishedTasks    []ITask
	mu               sync.Mutex
}

func NewTaskPool() *TaskPool {
	pool := &TaskPool{
		finishedTasks: make([]ITask, 0),
	}
	return pool
}

func (pool *TaskPool) PushTask(task ITask) {
	t := &Task{
		task:       task,
		pool:       pool,
		activeChan: make(chan ITask),
	}
	pool.pendingTaskCount.Add(1)
	go t.run()
	t.attachTask(task)
}

func (pool *TaskPool) PendingTaskCount() int {
	return int(pool.pendingTaskCount.Load())
}

func (pool *TaskPool) FinishTaskCount() int {
	return len(pool.finishedTasks)
}

func (pool *TaskPool) pushFinishedTask(task ITask) {
	pool.mu.Lock()
	defer pool.mu.Unlock()
	pool.finishedTasks = append(pool.finishedTasks, task)
}

func (pool *TaskPool) getFinishedTask() ITask {
	pool.mu.Lock()
	defer pool.mu.Unlock()
	if len(pool.finishedTasks) == 0 {
		return nil
	}
	task := pool.finishedTasks[0]
	pool.finishedTasks = pool.finishedTasks[1:]
	return task
}

func (pool *TaskPool) Loop() {
	task := pool.getFinishedTask()
	if task != nil {
		task.Finish()
	}
}
