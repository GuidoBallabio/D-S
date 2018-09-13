package account

import (
	"sync"
)

// Ledger is synchronized account map
type Ledger struct {
	Accounts map[string]int
	lock     sync.RWMutex
}

// NewLedger is a constructor of ledgers
func NewLedger() *Ledger {
	var l Ledger
	l.Accounts = make(map[string]int, 10)
	return &l
}

// Transaction is method of ledger that applias a transaction to itself
func (l *Ledger) Transaction(t *Transaction) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.Accounts[t.From] -= t.Amount
	l.Accounts[t.To] += t.Amount
}
