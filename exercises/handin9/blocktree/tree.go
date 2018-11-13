package blocktree

import (
	b "bytes"
	"math/big"
	"time"

	. "../account"
)

// NodeSet is the set of all nodes indexed by their hash
var nodeSet = map[nodeHash]*Node{}

// Genesis is the root of the tree
var genesis nodeHash

//// Current State
// Head is the node considered the current head of the chain, so head == leafs[0] until there is a rollback
var head nodeHash

// Ledger: current state given by head
var ledger *Ledger

// leafs is the array of the leafs of the tree sorted for descending longest path to the root
var leafs []nodeHash

// Hardness is the number from which derives the probability of winning
var hardness *big.Int

// SlotLength is the time duration of the Slot
var slotLength = 1 * time.Second

// NewTree create a tree with the given root (genesis) block
func NewTree(initTrans []Transaction) {
	hardness = new(big.Int).Exp(big.NewInt(2), big.NewInt(255-3), nil)
	slotLength = 1 * time.Second

	gen := &Node{
		Slot:         0,
		CreatedStake: initTrans}
	genesis = hashNode(gen)
	nodeSet[genesis] = gen

	head = genesis
	leafs = []nodeHash{genesis}
	updateLedger()
}

// Partecipating returns true if the value of the draw on the local machine is higher than the Hardness
func Partecipating(node *Node) bool {
	return node.valueOfDraw().Cmp(hardness) == 1
}

// CompareValueOfNodes returns true if n1 wins over n2
func CompareValueOfNodes(n1, n2 *Node) bool {
	cmp := n1.valueOfDraw().Cmp(n2.valueOfDraw())

	if cmp == 0 {
		//Give advantage to the alphabetically sorted pubKeys
		return n1.Peer < n2.Peer
	}

	return cmp == 0
}

func getStake(peer string) int64 {
	return int64(ledger.GetBalance(peer))
}

// UpdateLedger recreates ledger up to the current head
func updateLedger() {
	path, found := pathFromTo(head, leafs[0])

	if !found {
		path, _ = pathFromTo(genesis, leafs[0])
		ledger = NewLedger()

		for i, nh := range path {
			applyAllTransactions(nh.getNode())
		}

		return
	}

	// skip head itself from being reapplied
	for i, nh := range path[1:] {
		applyAllTransactions(nh.getNode())
	}
}

// PathFromTo returns the path between two nodes if it exists otherwise (nil, false)
func pathFromTo(from, to nodeHash) ([]nodeHash, bool) {
	path := []nodeHash{}

	end := false
	found := false

	for nh := to; !end && !found; nh = nh.getNode().Parent {
		path = append([]nodeHash{nh}, path...)

		end = b.Equal(nh[:], genesis[:])
		found = b.Equal(nh[:], from[:])
	}

	if found {
		return append([]nodeHash{from}, path...), true
	}

	return nil, false
}

func applyAllTransactions(node *Node) {
	for _, id := range node.TransList {
		ledger.Transaction(inTransit.GetTransaction(id))
	}
}
