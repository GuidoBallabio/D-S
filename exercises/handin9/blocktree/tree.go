package blocktree

import (
	"math/big"
	"sync"
	"time"

	. "../account"
)

// Tree is the whole blockchain struct
type Tree struct {
	// NodeSet is the set of all nodes indexed by their hash
	nodeSet map[nodeHash]*Node

	// Genesis is the root of the tree
	genesis nodeHash

	// Head is the node considered the current top of the chain for the ledger, so head == leafs[0] until there is a rollback
	head nodeHash

	// Ledger: current state given by head
	ledger *Ledger

	// leafs is the array of the leafs of the tree sorted for descending longest path to the root
	leafs []nodeHash

	// Hardness is the number from which derives the probability of winning
	hardness *big.Int

	// SlotLength is the time duration of the Slot
	slotLength time.Duration

	// lock for synchronization
	lock sync.RWMutex
}

// NewTree create a tree with the given root (genesis) block
func NewTree(initTrans []Transaction) *Tree {
	gen := &Node{
		Seed:         42, //for reproducibility is fixed in development
		Slot:         0,
		CreatedStake: initTrans}

	genHash := gen.hash()

	tree := &Tree{
		nodeSet:    map[nodeHash]*Node{},
		genesis:    genHash,
		head:       genHash,
		ledger:     NewLedger(),
		leafs:      []nodeHash{genHash},
		hardness:   new(big.Int).Exp(big.NewInt(2), big.NewInt(255-3), nil),
		slotLength: 1 * time.Second}

	tree.nodeSet[tree.genesis] = gen

	tree.updateLedger()

	return tree
}

// Partecipating returns true if the value of the draw on the local machine is higher than the Hardness
func (t *Tree) Partecipating(node *Node) bool {
	return node.valueOfDraw(t).Cmp(t.hardness) == 1
}

// CompareValueOfNodes returns true if n1 wins over n2
func (t *Tree) CompareValueOfNodes(n1, n2 *Node) bool {
	cmp := n1.valueOfDraw(t).Cmp(n2.valueOfDraw(t))

	if cmp == 0 {
		//Give advantage to the alphabetically sorted pubKeys
		return n1.Peer < n2.Peer
	}

	return cmp == 1
}

// GetSeed return the current seed
func (t *Tree) GetSeed() uint64 { //maybe needs locks
	return t.getNode(t.head).Seed
}

// CheckIsNext returns true if the node can be considered for addition false if it could be a future one
func (t *Tree) CheckIsNext(n *Node) bool {
	return n.getParent(t) != nil
}

// ConsiderLeaf tries to add a node to the tree as leaf (hence should be the winner)
// and return true if succeeds (the node should be discarded otherwise)
func (t *Tree) ConsiderLeaf(n *Node) bool {
	// Verify its consistency

	//// younger than parent
	if n.Slot <= n.getParent(t).Slot {
		return false
	}

	t.addLeaf(n)
	t.updateLedger()

	return true

}

// AddLeaf adds node to the correct position to the tree sorting the leafs as well
func (t *Tree) addLeaf(n *Node) {
	nh := n.hash()
	ph := nh.getParent(t)

	t.nodeSet[nh] = n

	found := false
	index := 0

	// check if parent is one of the leafs
	for i, l := range t.leafs {
		if eqH(ph, l) {
			found = true
			index = i
			break
		}
	}

	if found { // if parent is one of the leafs just replace it with ph
		t.leafs[index] = nh
		//sort comparing to the next one
		if index > 0 && t.compareWeight(t.leafs[index], t.leafs[index-1]) {
			t.leafs[index-1], t.leafs[index] = t.leafs[index], t.leafs[index-1]
		}
		return
	}

	////// if it is not, add a new leaf a the correct sorted position

	// find correct position (could be binary search but too much effort)
	for i := range t.leafs {
		if t.compareWeight(nh, t.leafs[i]) {
			index = i
			break
		}
	}

	if index == 0 { //new node is the most important
		t.leafs = append([]nodeHash{nh}, t.leafs...)
		return
	}

	if index == len(t.leafs) { // new leaf has the lowest pathWeight
		t.leafs = append(t.leafs, nh)
		return
	}

	t.leafs = append(t.leafs, nodeHash{})
	copy(t.leafs[index+1:], t.leafs[index:])
	t.leafs[index] = nh
}

// CompareWeight decides which node (should be leaf) has higher priority (must be in the tree)
func (t *Tree) compareWeight(nh1, nh2 nodeHash) bool {
	len1 := t.pathLenght(nh1)
	len2 := t.pathLenght(nh2)

	if len1 > len2 {
		return true
	}

	if len1 == len2 && t.CompareValueOfNodes(t.nodeSet[nh1], t.nodeSet[nh2]) {
		return true
	}

	return false
}

// PathLenght of a node given that its in the tree already
func (t *Tree) pathLenght(nh nodeHash) uint64 {
	path, _ := t.pathFromTo(t.genesis, nh)
	return uint64(len(path))
}

func (t *Tree) getStake(peer string) int64 { //maybe needs locks
	return int64(t.ledger.GetBalance(peer))
}

// UpdateLedger recreates ledger up to the current head
func (t *Tree) updateLedger() {
	path, found := t.pathFromTo(t.head, t.leafs[0])

	if !found {
		path, _ = t.pathFromTo(t.genesis, t.leafs[0])
		t.ledger = NewLedger()

		for _, nh := range path {
			t.applyAllTransactions(nh.getNode(t))
		}

	} else {
		// skip head itself from being reapplied
		for _, nh := range path[1:] {
			t.applyAllTransactions(nh.getNode(t))
		}
	}

	t.head = t.leafs[0]
}

// PathFromTo returns the path between two nodes if it exists otherwise (nil, false)
func (t *Tree) pathFromTo(from, to nodeHash) ([]nodeHash, bool) {
	path := []nodeHash{}

	end := false
	found := false

	for nh := to; !end && !found; nh = nh.getParent(t) {
		path = append([]nodeHash{nh}, path...)

		end = eqH(nh, t.genesis)
		found = eqH(nh, from)
	}

	// found from or from == genesis hence found from
	if found || (found == end) {
		return append([]nodeHash{from}, path...), true
	}

	return nil, false
}

// ApplyAllTransactions applies a node to the ledger
func (t *Tree) applyAllTransactions(node *Node) {
	for _, tran := range node.TransList {
		t.ledger.Transaction(tran)
	}
	for _, tran := range node.CreatedStake {
		t.ledger.Transaction(tran)
	}
}

// GetNode gets a node given its hash
func (t *Tree) getNode(nh nodeHash) *Node {
	return nh.getNode(t)
}

// GetNode gets a node given its hash
func (t *Tree) getParent(nh nodeHash) nodeHash {
	return nh.getNode(t).Parent
}
