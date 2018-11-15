package services

import (
	"fmt"

	heap "github.com/emirpasic/gods/trees/binaryheap"
	"github.com/emirpasic/gods/utils"

	. "../account"
)

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