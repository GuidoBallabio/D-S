package services

import (
	"fmt"
	"sync"

	. "../account"
)

// Wg is the waitgroup for the services
var Wg sync.WaitGroup

var ledger = NewLedger()
var past = NewPastMap()
var inTransit = NewTransactionMap()

// ProcessTransactions handles the trasaction recieved
func ProcessTransactions(listenCh <-chan SignedTransaction, sequencerCh chan<- Transaction, quitCh <-chan struct{}) {
	defer Wg.Done()

	for {
		select {
		case st := <-listenCh:
			if t := st.ExtractTransaction(); !isOld(t) && isVerified(st) && ledger.CheckBalance(t) {
				inTransit.AddTransaction(t)
				past.AddPast(t, true)
				if CheckIfSequencer() {
					sequencerCh <- t
				}
				broadcast(st)
			}
		case <-quitCh:
			return //Done
		}
	}
}

func isOld(t Transaction) bool {
	if val, found := past.GetPast(t); found && val {
		return true
	}
	return false
}

func isVerified(st SignedTransaction) bool {
	return st.VerifyTransaction() && st.Amount > 0
}

func attachNextID(t Transaction) Transaction {
	t.ID = fmt.Sprintf("%d-%s", past.GetPastLength(), LocalPeer.GetAddress())
	past.AddPast(t, false)
	return t
}

func broadcast(st SignedTransaction) {
	var w WhatType = st
	for enc := range PeerList.IterEnc() {
		enc.Encode(&w)
	}
}
