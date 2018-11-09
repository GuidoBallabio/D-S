package account

import (
	"fmt"
	"sync"
)

// PastMap is synchronized trasaction map
type PastMap struct {
	data map[string]bool
	lock sync.RWMutex
}

// NewPastMap is a constructor of transaction maps
func NewPastMap() *PastMap {
	var l PastMap
	l.data = map[string]bool{}
	return &l
}

// GetPast is method of that retireves the transaction's past given the ID
func (l *PastMap) GetPast(t Transaction) (bool, bool) {
	l.lock.RLock()
	defer l.lock.RUnlock()

	val, found := l.data[t.ID]

	return val, found
}

// AddPast is method of that adds the transaction to the map
func (l *PastMap) AddPast(t Transaction, b bool) {
	l.lock.Lock()
	defer l.lock.Unlock()

	l.data[t.ID] = b
}

// GetPastLength is method of that retireves the transaction's past legnth
func (l *PastMap) GetPastLength() int {
	l.lock.RLock()
	defer l.lock.RUnlock()

	return len(l.data)
}

func (l *PastMap) String() string {
	l.lock.RLock()
	defer l.lock.RUnlock()

	return fmt.Sprintln(l.data)
}
