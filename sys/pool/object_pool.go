package pool

import (
	col "nemo/sys/collections"
	"sync"
)

type ObjectPool struct {
	sync.RWMutex

	usedObjs map[any]struct{}
	freeObjs *col.Queue
}

func NewObjectPool() *ObjectPool {

	class := new(ObjectPool)
	class.usedObjs = make(map[any]struct{}, 1024)
	class.freeObjs = col.NewQueue(1024)

	return class
}

// Create object to pool.
func (class *ObjectPool) Create(obj any) {
	class.Lock()
	defer class.Unlock()

	class.freeObjs.Enqueue(obj)
}

// Get a free object from pool.
func (class *ObjectPool) Get() any {
	class.RLock()
	defer class.RUnlock()

	obj := class.freeObjs.Dequeue()
	class.usedObjs[obj] = struct{}{}

	return obj
}

// Free an object to pool.
func (class *ObjectPool) Free(obj any) {
	class.Lock()
	defer class.Unlock()

	delete(class.usedObjs, obj)
	class.freeObjs.Enqueue(obj)
}

// Bind some value to used pool object.
func (class *ObjectPool) Bind(obj any, val struct{}) {
	class.Lock()
	defer class.Unlock()
	class.usedObjs[obj] = val
}

// Range all objects
func (class *ObjectPool) Range(f func(any)) {
	class.RLock()
	defer class.RUnlock()

	for k := range class.usedObjs {
		f(k)
	}
	class.freeObjs.Range(f)
}

func (class *ObjectPool) UsedRange(f func(any)) {
	class.RLock()
	defer class.RUnlock()

	for k := range class.usedObjs {
		f(k)
	}
}

// UsedCount Return the count of current used.
func (class *ObjectPool) UsedCount() int {
	class.RLock()
	defer class.RUnlock()
	return len(class.usedObjs)
}

// FreeCount Return the count of current free.
func (class *ObjectPool) FreeCount() int {
	class.RLock()
	defer class.RUnlock()
	return class.freeObjs.Count()
}
