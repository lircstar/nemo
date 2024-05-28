package util

import (
	"sync"
)

type SafeMap struct {
	mu sync.RWMutex
	m  map[any]any
}

func NewSafeMap(maxCount int) *SafeMap {
	return &SafeMap{
		m: make(map[any]any, maxCount),
	}
}

func (sm *SafeMap) Store(key any, value any) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.m[key] = value
}

func (sm *SafeMap) Load(key any) (value any, ok bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	value, ok = sm.m[key]
	return
}

func (sm *SafeMap) Delete(key any) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.m, key)
}

func (sm *SafeMap) Len() int {
	return len(sm.m)
}

func (sm *SafeMap) Range(f func(key any, value any) bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	for k, v := range sm.m {
		if !f(k, v) {
			break
		}
	}
}
