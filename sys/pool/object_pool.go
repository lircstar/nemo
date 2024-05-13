package pool

import (
	col "nemo/sys/collections"
	"sync"
)

type ObjectPool struct {
	sync.Mutex

	usedObjs map[interface{}]struct{}
	freeObjs *col.Queue
}

func NewObjectPool() *ObjectPool {

	class := new(ObjectPool)
	class.usedObjs = make(map[interface{}]struct{}, 1024)
	class.freeObjs = col.NewQueue(1024)

	return class
}

// Create object to pool.
func (class *ObjectPool) Create(obj interface{}) {
	class.Lock()
	defer class.Unlock()

	class.freeObjs.Enqueue(obj)
}

// Get a free object from pool.
func (class *ObjectPool) Get() interface{} {
	class.Lock()
	defer class.Unlock()

	obj := class.freeObjs.Dequeue()
	class.usedObjs[obj] = struct{}{}

	return obj
}

// Free an object to pool.
func (class *ObjectPool) Free(obj interface{}) {
	class.Lock()
	defer class.Unlock()

	delete(class.usedObjs, obj)
	class.freeObjs.Enqueue(obj)
}

// Bind some value to used pool object.
func (class *ObjectPool) Bind(obj interface{}, val struct{}) {
	class.Lock()
	defer class.Unlock()
	class.usedObjs[obj] = val
}

// Free all objects
func (class *ObjectPool) Range(f func(interface{})) {
	class.Lock()
	defer class.Unlock()

	for k := range class.usedObjs {
		f(k)
	}
	class.freeObjs.Range(f)
}

func (class *ObjectPool) UsedRange(f func(interface{})) {
	class.Lock()
	defer class.Unlock()

	for k := range class.usedObjs {
		f(k)
	}
}

// Return the count of current used.
func (class *ObjectPool) UsedCount() int {
	class.Lock()
	defer class.Unlock()
	return len(class.usedObjs)
}

// Return the count of current free.
func (class *ObjectPool) FreeCount() int {
	class.Lock()
	defer class.Unlock()
	return class.freeObjs.Count()
}
