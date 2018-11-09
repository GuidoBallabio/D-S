package account

import (
	"fmt"
	"sync"
)

// TransactionMap is synchronized trasaction map
type TransactionMap struct {
	data map[string]Transaction
	lock sync.RWMutex
}

// NewTransactionMap is a constructor of transaction maps
func NewTransactionMap() *TransactionMap {
	var l TransactionMap
	l.data = make(map[string]Transaction, 1)
	return &l
}

// GetTransaction is method of that retireves the transaction given the ID
func (l *TransactionMap) GetTransaction(id string) Transaction {
	l.lock.RLock()
	defer l.lock.RUnlock()

	return l.data[id]
}

// AddTransaction is method of that adds the transaction to the map
func (l *TransactionMap) AddTransaction(t Transaction) {
	l.lock.Lock()
	defer l.lock.Unlock()

	l.data[t.ID] = t
}

func (l *TransactionMap) String() string {
	l.lock.RLock()
	defer l.lock.RUnlock()

	return fmt.Sprintln(l.data)
}
