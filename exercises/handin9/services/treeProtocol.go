package services

import (
	"fmt"
	"time"

	. "../account"
	"../aesrsa"
	bt "../blocktree"
)

// Tree is the blockchain tree
var Tree *bt.Tree

// ProcessNodes implements the tree protocol
func ProcessNodes(sequencerCh <-chan Transaction, blockCh <-chan bt.SignedNode, keys *aesrsa.RSAKeyPair, quitCh <-chan struct{}) {
	defer Wg.Done()

	ticker := time.NewTicker(Tree.SlotLength)
	seq := make([]string, 0)
	var winner *bt.Node
	nodeOfSlot := map[*bt.Node]struct{}{}

	for {
		select {
		case <-ticker.C:
			Tree.IncrementSlot()
			nodeOfSlot = map[*bt.Node]struct{}{}

			// use winner for currentSlot-1
			if winner != nil {
				Tree.ConsiderLeaf(winner)
				fmt.Println(Tree.GetLedger())
			}

			// make own node for current slot (just ended)
			if len(seq[:]) > 0 {
				n := bt.NewNode(Tree.GetSeed(), Tree.GetCurrentSlot(), seq, keys, Tree.GetHead())
				sn := bt.NewSignedNode(*n, keys.Private)
				go broadcastNode(*sn)
				seq = make([]string, 0)

				winner = n
			} else {
				winner = nil
			}
		case t := <-sequencerCh:
			if Tree.ConsiderTransaction(t, seq) {
				seq = append(seq, t.ID)
			}
		case sn := <-blockCh:
			if n := &sn.Node; alreadySeen(n, nodeOfSlot) && Tree.CheckIsNext(n) && sn.VerifyNode() {
				nodeOfSlot[n] = struct{}{}
				if winner == nil || Tree.CompareValueOfNodes(n, winner) {
					winner = n
				}
				go broadcastNode(sn)
			}
		case <-quitCh:
			return //Done
		}
	}
}

func alreadySeen(n *bt.Node, nodeOfSlot map[*bt.Node]struct{}) bool {
	_, found := nodeOfSlot[n]
	return found
}

func broadcastNode(sn bt.SignedNode) {
	Wg.Add(1)
	defer Wg.Done()

	var w WhatType = sn
	for enc := range PeerList.IterEnc() {
		enc.Encode(&w)
	}
}

/*

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
*/
