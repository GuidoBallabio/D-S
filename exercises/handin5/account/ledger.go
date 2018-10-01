package account

import (
	"errors"
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

// Transaction is method of ledger that applies a transaction to itself
func (l *Ledger) Transaction(t Transaction) {
	l.lock.Lock()
	defer l.lock.Unlock()

	l.clock++
	l.Accounts[t.From] -= t.Amount
	l.Accounts[t.To] += t.Amount
}

// TransactionWithBalanceCheck applies a transaction if not out of balance
func (l *Ledger) TransactionWithBalanceCheck(t Transaction) error {
	l.lock.Lock()
	defer l.lock.Unlock()

	val, found := l.Accounts[t.From]

	if !found {
		l.clock++
		l.Accounts[t.From] = t.Amount
		return nil
	}

	if val < t.Amount {
		return errors.New("Overdrwaing from balance not allowed")
	}

	l.clock++
	l.Accounts[t.From] -= t.Amount
	l.Accounts[t.To] += t.Amount

	return nil
}

// GetClock return the ledger clock
func (l *Ledger) GetClock() uint64 {
	l.lock.RLock()
	defer l.lock.RUnlock()

	return l.clock
}

func (l *Ledger) String() string {
	l.lock.RLock()
	defer l.lock.RUnlock()

	return fmt.Sprintln(l.Accounts)
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
