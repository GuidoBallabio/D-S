package blocktree

import (
	"crypto/sha256"
	"encoding/json"
	"math/big"

	. "../account"
	"../aesrsa"
)

// Node is a node of the tree in the blockchain
type Node struct {
	Seed         uint64
	Slot         uint64
	Peer         string
	Draw         []byte
	CreatedStake []Transaction
	TransList    []Transaction
	Parent       nodeHash
}

// NewNode given slot number and transactions
func NewNode(seed, slot uint64, transList []Transaction, keys aesrsa.RSAKeyPair, parent *Node) *Node {
	return &Node{
		Seed:      seed,
		Slot:      slot,
		Peer:      aesrsa.KeyToString(keys.Public),
		Draw:      getDraw(slot, seed, keys.Private),
		TransList: transList,
		Parent:    parent.hash()}
}

// GetParent returns the parent of the node
func (n *Node) getParent(t *Tree) *Node { //maybe needs locks
	val, _ := t.nodeSet[n.Parent]
	return val
}

func (n *Node) valueOfDraw(t *Tree) *big.Int {
	var val big.Int

	json1, err := json.Marshal(n.Slot)
	check(err)
	json2, err := json.Marshal(n.Seed)
	check(err)
	json3, err := json.Marshal(n.Draw)
	check(err)
	json4, err := json.Marshal(n.Peer)
	check(err)
	json := append(json1, json2...)
	json = append(json, json3...)
	json = append(json, json4...)

	hash := sha256.Sum256(json)

	hashInt := new(big.Int).SetBytes(hash[:])

	return val.Mul(hashInt, big.NewInt(t.getStake(n.Peer)))
}

func (n *Node) hash() nodeHash {
	return hashNode(n)
}

// utils

func getDraw(slot, seed uint64, sk aesrsa.RSAKey) []byte {
	json1, err := json.Marshal(slot)
	check(err)
	json2, err := json.Marshal(seed)
	check(err)

	return aesrsa.SignRSA(append(json1, json2...), sk)
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
