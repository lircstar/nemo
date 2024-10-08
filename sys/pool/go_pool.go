package pool

import (
	"fmt"
	"sync"
	"time"
)

type IRoutineTask interface {
	Run()
	Finish()
}

type TaskRoutine struct {
	pool *RoutinePool
	task IRoutineTask
}

func newTaskRoutine(pool *RoutinePool) *TaskRoutine {
	t := &TaskRoutine{pool: pool}
	go t.Run()
	return t
}

func (t *TaskRoutine) AttachTask(task IRoutineTask) {
	t.task = task
	t.pool.activeRoutines <- t
}

func (t *TaskRoutine) Run() {
	for {
		task := <-t.pool.activeRoutines
		task.task.Run()
		t.pool.pushFinishedTask(task.task)
		t.pool.writeThreadLog()

		task.task = t.pool.getPendingTask()
		if task.task == nil {
			break
		}
	}
	t.pool.pushFreeRoutine(t)
}

type RoutinePool struct {
	routineCount    int
	freeRoutines    chan *TaskRoutine
	activeRoutines  chan *TaskRoutine
	pendingRoutines []IRoutineTask
	finishedTasks   []IRoutineTask
	printLog        bool
	mu              sync.Mutex
}

func NewRoutinePool(routineCount int) *RoutinePool {
	pool := &RoutinePool{
		routineCount:    routineCount,
		freeRoutines:    make(chan *TaskRoutine, routineCount),
		activeRoutines:  make(chan *TaskRoutine, routineCount),
		pendingRoutines: make([]IRoutineTask, 0),
		finishedTasks:   make([]IRoutineTask, 0),
	}

	for i := 0; i < routineCount; i++ {
		thread := newTaskRoutine(pool)
		pool.freeRoutines <- thread
	}

	return pool
}

func (pool *RoutinePool) PushTask(task IRoutineTask) {
	select {
	case thread := <-pool.freeRoutines:
		thread.AttachTask(task)
	default:
		pool.mu.Lock()
		pool.pendingRoutines = append(pool.pendingRoutines, task)
		pool.mu.Unlock()
	}
}

func (pool *RoutinePool) pushFreeRoutine(thread *TaskRoutine) {
	pool.freeRoutines <- thread
}

func (pool *RoutinePool) getPendingTask() IRoutineTask {
	pool.mu.Lock()
	defer pool.mu.Unlock()
	if len(pool.pendingRoutines) == 0 {
		return nil
	}
	task := pool.pendingRoutines[0]
	pool.pendingRoutines = pool.pendingRoutines[1:]
	return task
}

func (pool *RoutinePool) pushFinishedTask(task IRoutineTask) {
	pool.mu.Lock()
	pool.finishedTasks = append(pool.finishedTasks, task)
	pool.mu.Unlock()
}

func (pool *RoutinePool) getFinishedTask() IRoutineTask {
	pool.mu.Lock()
	defer pool.mu.Unlock()
	if len(pool.finishedTasks) == 0 {
		return nil
	}
	task := pool.finishedTasks[0]
	pool.finishedTasks = pool.finishedTasks[1:]
	return task
}

func (pool *RoutinePool) writeThreadLog() {
	if pool.printLog {
		now := time.Now().Unix()
		if time.Now().Unix()-now >= 10 {
			now = time.Now().Unix()
			pool.mu.Lock()
			pendingCount := len(pool.pendingRoutines)
			finishedCount := len(pool.finishedTasks)
			freeCount := pool.routineCount - len(pool.freeRoutines)
			pool.mu.Unlock()

			fmt.Printf("POOLID:%p, PendingTask:%d, FinishedTask:%d, FreeThread:%d\n", pool, pendingCount, finishedCount, freeCount)
		}
	}
}

func (pool *RoutinePool) SetPrintLog(set bool) {
	pool.printLog = set
}

func (pool *RoutinePool) GetPendingTaskCount() int {
	pool.mu.Lock()
	defer pool.mu.Unlock()
	return len(pool.pendingRoutines)
}

func (pool *RoutinePool) GetFinishedTaskCount() int {
	pool.mu.Lock()
	defer pool.mu.Unlock()
	return len(pool.finishedTasks)
}

func (pool *RoutinePool) Loop() {
	for {
		task := pool.getFinishedTask()
		if task != nil {
			task.Finish()
		}
	}
}

//func main() {
//	// Example usage
//	pool := NewRoutinePool(5)
//	pool.SetPrintLog(true)
//	go pool.writeThreadLog()
//
//	// Add tasks to the pool
//	for i := 0; i < 10; i++ {
//		task := &ExampleTask{id: i}
//		pool.PushTask(task)
//	}
//
//	// Run the loop to process finished tasks
//	pool.Loop()
//
//	// Wait for a while to let tasks complete
//	time.Sleep(30 * time.Second)
//}
//
//type ExampleTask struct {
//	id int
//}
//
//func (t *ExampleTask) Run() {
//	fmt.Printf("Running task %d\n", t.id)
//	time.Sleep(1 * time.Second)
//}
//
//func (t *ExampleTask) Finish() {
//	fmt.Printf("Finished task %d\n", t.id)
//}
