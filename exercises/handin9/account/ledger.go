package account

import (
	"fmt"
	"sort"
	"strconv"
	"sync"
)

// Ledger is synchronized account map
type Ledger struct {
	Accounts map[string]uint64
	lock     sync.RWMutex
}

// NewLedger is a constructor of ledgers
func NewLedger() *Ledger {
	var l Ledger
	l.Accounts = make(map[string]uint64, 1)
	return &l
}

// Transaction is method of ledger that applies a transaction to itself
func (l *Ledger) Transaction(t Transaction) {
	l.lock.Lock()
	defer l.lock.Unlock()

	_, found := l.Accounts[t.From]

	if l.checkBalance(t) {
		if !found {
			return
		}

		l.Accounts[t.From] -= t.Amount
		l.Accounts[t.To] += t.Amount

	}
}

// checkBalance confirms a transacrtion won't put someone out of balance
// without locks as private for use inside other locked func
func (l *Ledger) checkBalance(t Transaction) bool {

	val, found := l.Accounts[t.From]

	if !found {
		return true
	}

	if val < t.Amount {
		return false
	}

	return true
}

// CheckBalance confirms a transacrtion won't put someone out of balance
func (l *Ledger) CheckBalance(t Transaction) bool {
	l.lock.RLock()
	defer l.lock.RUnlock()

	return l.checkBalance(t)
}

// GetBalance returns the balance of a peer
func (l *Ledger) GetBalance(peer string) uint64 {
	l.lock.RLock()
	defer l.lock.RUnlock()

	return l.Accounts[peer]
}

// AddToBalance creates money giving it to a peer
func (l *Ledger) AddToBalance(peer string, amount uint64) {
	l.lock.Lock()
	defer l.lock.Unlock()

	l.Accounts[peer] += amount

}

func (l *Ledger) String() string {
	l.lock.RLock()
	defer l.lock.RUnlock()

	var s = ""
	for key, value := range l.Accounts {
		s = s + fmt.Sprintf("Key: "+key[20:29]+" | Value: "+strconv.Itoa(int(value))+"\n")
	}

	return s
}

// GetSortedKeys returns a sorted list of keys
func (l *Ledger) GetSortedKeys() []string {
	l.lock.RLock()
	defer l.lock.RUnlock()

	type led = struct {
		Key    string
		Amount uint64
	}

	list := []led{}

	for key, value := range l.Accounts {
		ledg := led{
			Key:    key,
			Amount: value}
		list = append(list, ledg)
	}

	sort.SliceStable(list, func(i, j int) bool {
		return list[i].Amount < list[j].Amount
	})

	var array []string

	for i, j := range list {
		array[i] = j.Key
	}

	return array
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
