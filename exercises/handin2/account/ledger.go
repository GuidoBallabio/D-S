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

// Transaction is an atomic operation on a ledger
type Transaction struct {
	ID     string
	From   string
	To     string
	Amount int
}

// NewTransaction is a constructor of transactions
func NewTransaction(ID, From, To string, Amount int) *Transaction {
	return &Transaction{
		ID:     ID,
		From:   From,
		To:     To,
		Amount: Amount}
}

// Transaction is method of ledger that applias a transaction to itself
func (l *Ledger) Transaction(t *Transaction) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.Accounts[t.From] -= t.Amount
	l.Accounts[t.To] += t.Amount
}
