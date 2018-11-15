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

// GetTransaction is method that retireves the transaction given the ID
func (l *TransactionMap) GetTransaction(id string) (Transaction, bool) {
	l.lock.RLock()
	defer l.lock.RUnlock()

	val, found := l.data[id]
	return val, found
}

// SetTransaction is method that adds the transaction to the map
func (l *TransactionMap) SetTransaction(t Transaction) {
	l.lock.Lock()
	defer l.lock.Unlock()

	l.data[t.ID] = t
}

// RemoveID is method that removes the transaction from the map given the id
func (l *TransactionMap) RemoveID(id string) {
	l.lock.Lock()
	defer l.lock.Unlock()

	delete(l.data, id)
}

// RemoveTransaction is method that removes the transaction from the map
func (l *TransactionMap) RemoveTransaction(t Transaction) {
	l.lock.Lock()
	defer l.lock.Unlock()

	delete(l.data, t.ID)
}

// Reset is method that removes every transaction from the map
func (l *TransactionMap) Reset() {
	l.lock.Lock()
	defer l.lock.Unlock()

	l.data = make(map[string]Transaction, 1)
}

// Iter is method that allows to iterate on the map
func (l *TransactionMap) Iter() <-chan Transaction { // not synchronyzed enough for union
	c := make(chan Transaction)

	f := func() {
		l.lock.RLock()
		defer l.lock.RUnlock()
		for _, value := range l.data {
			c <- value
		}
		close(c)
	}
	go f()

	return c
}

// TransferAll does l2 = l2 U l and l.Reset() atomically
func (l *TransactionMap) TransferAll(l2 *TransactionMap) {
	l.lock.Lock()
	l2.lock.Lock()
	defer l2.lock.Unlock()
	defer l.lock.Unlock()

	for k, v := range l.data {
		l2.data[k] = v
	}

	l.data = make(map[string]Transaction, 1)
}

func (l *TransactionMap) String() string {
	l.lock.RLock()
	defer l.lock.RUnlock()

	return fmt.Sprintln(l.data)
}
