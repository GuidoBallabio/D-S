package services

import (
	"fmt"
	"time"

	. "../account"
	"../aesrsa"
	bt "../blocktree"
)

var Tree *Tree

// BeSequencer add the beheaviour of a sequencer to the peer
func BeSequencer(sequencerCh <-chan Transaction, blockCh chan<- bt.SignedNode, keys aesrsa.RSAKeyPair, quitCh <-chan struct{}) {
	defer Wg.Done()

	fmt.Println("You are the Sequencer")

	ticker := time.NewTicker(Tree.SlotLength)

	for {
		seq := make([]Transaction, 0)
		endBlock := false
		for !endBlock {
			select {
			case <-ticker.C:
				if len(seq[:]) > 0 {
					n := bt.NewNode(t.GetSeed(), t.CurrentSlot, seq, keys, t.GetHead())
					t.IncrementSlot()
					sn := bt.NewSignedNode(n, key.Private)
					broadcastNode(*sn)
					blockCh <- *sn
					endBlock = true
				}
			case t := <-sequencerCh: //remove ? usless already received map?
				if Tree.ConsiderTransaction(t) {
					seq = append(seq, t)
				}
			case <-quitCh:
				return //Done
			}
		}
	}
}

func broadcastNode(sn bt.SignedNode) {
	var w WhatType = sn
	for enc := range PeerList.IterEnc() {
		enc.Encode(&w)
	}
