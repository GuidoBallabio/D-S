package services

import (
	"fmt"

	heap "github.com/emirpasic/gods/trees/binaryheap"
	"github.com/emirpasic/gods/utils"

	. "../account"
)

// ProcessBlocks applys blocks of transactions to the ledger
func ProcessBlocks(blockCh <-chan SignedBlock, quitCh <-chan struct{}) {
	defer Wg.Done()

	comp := func(a, b interface{}) int {
		b1 := a.(Block)
		b2 := b.(Block)
		return utils.IntComparator(b1.Number, b2.Number)
	}

	pq := heap.NewWith(comp)
	defer pq.Clear()

	lastBlock := -1

	for {
		select {
		case sb := <-blockCh:
			if b := sb.ExtractBlock(); sb.VerifyBlock(sequencer) && isFuture(b, lastBlock) {
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
func isFuture(b Block, lastBlock int) bool {
	return b.Number >= lastBlock+1
}

// isNext tells if it's the next block to be processed
func isNext(b Block, lastBlock int) bool {
	return b.Number == lastBlock+1
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
