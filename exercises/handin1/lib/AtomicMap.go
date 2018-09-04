package lib

import (
	"sync"
)

// AtomicMap is a synchronized map[string]bool type
type AtomicMap struct {
	data   map[string]bool
	rwLock sync.RWMutex
}

// NewAtomicMap is the constructor of the AtmoicMap type, it just creates an empty synchronized map
func NewAtomicMap() *AtomicMap {
	var amap AtomicMap
	amap.data = make(map[string]bool)
	return &amap
}

// Get is the synchronized get
func (amap *AtomicMap) Get(key string) (bool, bool) {
	amap.rwLock.RLock()
	defer amap.rwLock.RUnlock()
	val, found := amap.data[key]
	return val, found
}

// Set is the synchronized set
func (amap *AtomicMap) Set(key string, val bool) {
	amap.rwLock.Lock()
	defer amap.rwLock.Unlock()
	amap.data[key] = val
}
