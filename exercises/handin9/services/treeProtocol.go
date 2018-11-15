package services

import (
	"fmt"
	"time"

	. "../account"
	"../aesrsa"
	bt "../blocktree"
)

var Tree *Tree

// processNodes implements the tree protocol
func processNodes(sequencerCh <-chan Transaction, blockCh chan<- bt.SignedNode, keys aesrsa.RSAKeyPair, quitCh <-chan struct{}) {
	defer Wg.Done()

	ticker := time.NewTicker(Tree.SlotLength)
	seq := make([]Transaction, 0)

	t.IncrementSlot()
	var winner *Node

	for {
		select {
			case <-ticker.C:
				if len(seq[:]) > 0 {
						n := bt.NewNode(t.GetSeed(), t.CurrentSlot, seq, keys, t.GetHead())
						t.IncrementSlot()
						sn := bt.NewSignedNode(n, key.Private)
						broadcastNode(*sn)
						seq == make([]Transaction, 0)

						winner = n
				}

				winner = nil

			case t := <-sequencerCh: 
				if Tree.ConsiderTransaction(t) {
					seq = append(seq, t)
				}
			case n := <- blockCh:
				if n := sn.Node; sn.VerifyNode() && Tree.CheckIfExist(n) {
					pq.Push(b)
					broadcastBlock(sb)
					lastBlock = applyAllValidBlocks(pq, lastBlock)
					fmt.Println(ledger) //TODO better print
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
}

// ProcessBlocks applys blocks of transactions to the ledger
func ProcessBlocks(blockCh <-chan SignedNode, quitCh <-chan struct{}) {
	defer Wg.Done()

	comp := func(a, b interface{}) int {
		n1 := a.(Node)
		n2 := b.(Node)
		return utils.IntComparator(b1.Number, b2.Number)
	}

	pq := heap.NewWith(comp)
	defer pq.Clear()

	lastBlock := -1

	for {
		select {
		case sn := <-blockCh:
			if n := sn.Node; sn.VerifyNode() && isFuture(n) {

				pq.Push(b)
				broadcastBlock(sb)
				lastBlock = applyAllValidBlocks(pq, lastBlock)
				fmt.Println(ledger) //TODO better print
			}
		case <-quitCh:
			return //Done
		}
	}

}

// isFuture tells if it's already been processed
func isFuture(n *Node) bool {
	return Tree.CheckIfExist(n)
}

// Applys every transaction from a block
func updateLedger(b Block) {
	for _, id := range b.TransList {
		ledger.Transaction(inTransit.GetTransaction(id))
	}
}

func applyAllValidBlocks(pq *heap.Heap, lastBlock int) int {
	if !pq.Empty() {
		tmp, full := pq.Peek()
		min := tmp.(Block)
		for full && isNext(min, lastBlock) {
			tmp, _ := pq.Pop()
			min = tmp.(Block)

			updateLedger(min)
			lastBlock++

			tmp, full = pq.Peek()
			if full {
				min = tmp.(Block)
			}
		}
	}

	return lastBlock
}

// broadcast a signed block
func broadcastBlock(sb SignedBlock) {
	var w WhatType = sb
	for enc := range PeerList.IterEnc() {
		enc.Encode(&w)
	}
}
