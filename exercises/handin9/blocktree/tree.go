package blocktree

import (
	"math/big"
	"sync"
	"time"

	. "../account"
)

// Tree is the whole blockchain struct
type Tree struct {
	//////////// ACTUAL TREE ////////////

	// Genesis is the root of the tree
	genesis nodeHash

	// NodeSet is the set of all nodes indexed by their hash
	nodeSet map[nodeHash]*Node

	// Leafs is the array of the leafs of the tree sorted for descending longest path to the root
	leafs []nodeHash

	// CurrentSlot is the current slot number
	currentSlot uint64

	//////////// STATE ////////////

	// Delivered transactions already accounted
	delivered *TransactionMap

	// Received transactions to be processes
	received *TransactionMap

	// Head is the node considered the current top of the chain for the ledger, so head == leafs[0] until there is a rollback
	head nodeHash

	// Ledger: current state given by head
	ledger *Ledger

	/////////// PARAMETERS ////////////

	// Hardness is the number from which derives the probability of winning
	hardness *big.Int

	// SlotLength is the time duration of the Slot
	SlotLength time.Duration

	// Reward is the reward for each node won
	reward uint64

	// Fee is the fee for each transaction payed to the peer who is responsible for its node
	fee uint64

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
		nodeSet:     map[nodeHash]*Node{},
		genesis:     genHash,
		leafs:       []nodeHash{genHash},
		currentSlot: gen.Slot,
		delivered:   NewTransactionMap(),
		received:    NewTransactionMap(),
		head:        genHash,
		ledger:      NewLedger(),
		hardness:    new(big.Int).Exp(big.NewInt(2), big.NewInt(255-3), nil),
		SlotLength:  1 * time.Second,
		reward:      10,
		fee:         1}

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

// GetHead returns
func (t *Tree) GetHead() *Node {
	return t.nodeSet[t.head]
}

// IncrementSlot let the next follow
func (t *Tree) IncrementSlot() {
	t.currentSlot++
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

	// add to tree
	t.addLeaf(n)
	// update state
	t.updateLedger()

	return true
}

// ConsiderTransaction adds it to the received set and returns if the transaction is suitable for the head of this local machine
func (t *Tree) ConsiderTransaction(tran Transaction, seq []string) bool {
	t.received.SetTransaction(tran)

	tmpLedger := t.ledger.Copy()

	for _, pTran := range seq {
		val, _ := t.received.GetTransaction(pTran)
		newTran, _ := t.deductFees(val)

		// Apply transaction
		t.ledger.Transaction(newTran)

	}

	return tmpLedger.CheckBalance(tran)
}

// GetLedger returns the current ledger status (to be printed)
func (t *Tree) GetLedger() string {
	return t.ledger.String()
}

// GetAccountNumbers return the list ok pubkeys in the ledger
func (t *Tree) GetAccountNumbers() []string {
	return t.ledger.GetSortedKeys()
}

// GetCurrentSlot returns the current slot number
func (t *Tree) GetCurrentSlot() uint64 {
	return t.currentSlot
}

// AddLeaf adds node to the correct position to the tree sorting the leafs as well
func (t *Tree) addLeaf(n *Node) {
	nh := n.hash()
	ph := n.Parent

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
		// New path from root
		path, _ = t.pathFromTo(t.genesis, t.leafs[0])
		// Recreate ledger
		t.ledger = NewLedger()

		// Reset delivered: delivered = empty, received = received U delivered
		t.delivered.TransferAll(t.received)

		for _, nh := range path {
			t.applyAllTransactions(nh.getNode(t))
		}

	} else {
		// proced on usual from head to new leaf, skip head itself from being reapplied
		for _, nh := range path[1:] {
			t.applyAllTransactions(nh.getNode(t))
		}
	}

	t.head = t.leafs[0]
}

// ApplyAllTransactions applies a node to the ledger and consider reward
func (t *Tree) applyAllTransactions(node *Node) {

	if node == t.nodeSet[t.genesis] {
		for _, tran := range node.CreatedStake {
			t.ledger.AddToBalance(tran.To, tran.Amount)
		}
		return
	}

	rewardPlusFees := t.reward

	for _, id := range node.TransList {
		tran, found := t.received.GetTransaction(id)
		if found {
			newTran, fee := t.deductFees(tran)

			// Apply transaction
			t.ledger.Transaction(newTran)
			rewardPlusFees += fee

			//Move from received to delivered
			t.received.RemoveID(id)
			t.delivered.SetTransaction(tran)
		} else {
			// wait and repeat? shouldn't happen
		}
	}

	t.ledger.AddToBalance(node.Peer, rewardPlusFees)
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

func (t *Tree) deductFees(tran Transaction) (Transaction, uint64) {
	tran.Amount -= t.fee

	return tran, t.fee
}

// GetNode gets a node given its hash
func (t *Tree) getNode(nh nodeHash) *Node {
	return nh.getNode(t)
}

// GetNode gets a node given its hash
func (t *Tree) getParent(nh nodeHash) nodeHash {
	return nh.getNode(t).Parent
}
