package blocktree

import (
	"math/big"
	"time"

	. "../account"
)

// Genesis is the root of the tree
var genesis *Node

// Ledger current state
var ledger *Ledger

// Head is the node considered the current head of the chain
var head *Node

// Hardness is the number from which derives the probability of winning
var hardness *big.Int

// SlotLength is the time duration of the Slot
var slotLength = 1 * time.Second

// NewTree create a tree with the given root (genesis) block
func NewTree(initTrans []Transaction) {
	hardness = new(big.Int).Exp(big.NewInt(2), big.NewInt(255-3), nil)
	slotLength = 1 * time.Second
	genesis = &Node{
		Slot:         0,
		CreatedStake: initTrans}
	head = genesis
	ledger = NewLedger()
}

// Partecipating returns true if the value of the draw on the local machine is higher than the Hardness
func Partecipating(node *Node) bool {
	return valueOfDraw(node).Cmp(hardness) == 1
}

// CompareValueOfNodes returns true if n1 wins over n2
func CompareValueOfNodes(n1, n2 *Node) bool {
	cmp := valueOfDraw(n1).Cmp(valueOfDraw(n2))

	if cmp == 0 {
		//Give advantage to the alphabetically sorted pubKeys
		return n1.Peer < n2.Peer
	}

	return cmp == 0
}
