package account

import (
	"fmt"
	"sync"
)

// Ledger is synchronized account map
type Ledger struct {
	Accounts map[string]int
	clock    uint64
	lock     sync.RWMutex
}

// NewLedger is a constructor of ledgers
func NewLedger() *Ledger {
	var l Ledger
	l.Accounts = make(map[string]int, 1)
	return &l
}

// Transaction is method of ledger that applias a transaction to itself
func (l *Ledger) Transaction(t Transaction) {
	l.lock.Lock()
	defer l.lock.Unlock()

	l.clock++
	l.Accounts[t.From] -= t.Amount
	l.Accounts[t.To] += t.Amount
}

// GetClock return the ledger clock
func (l *Ledger) GetClock() uint64 {
	l.lock.RLock()
	defer l.lock.RUnlock()

	return l.clock
}

func (l *Ledger) String() string {
	return fmt.Sprintln(l.Accounts)
}
