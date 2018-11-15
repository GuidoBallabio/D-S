package services

import (
	"fmt"
	"sync"

	. "../account"
)

// Wg is the waitgroup for all the services
var Wg sync.WaitGroup

var past = NewPastMap()

// ProcessTransactions handles the trasaction recieved
func ProcessTransactions(listenCh <-chan SignedTransaction, sequencerCh chan<- Transaction, quitCh <-chan struct{}) {
	defer Wg.Done()

	for {
		select {
		case st := <-listenCh:
			if t := st.ExtractTransaction(); !isOld(t) && isVerified(st) {
				past.AddPast(t, true)
				sequencerCh <- t
				go broadcast(st)
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
	Wg.Add(1)
	defer Wg.Done()

	var w WhatType = st
	for enc := range PeerList.IterEnc() {
		enc.Encode(&w)
	}
}
