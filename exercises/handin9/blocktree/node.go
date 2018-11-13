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
	TransList    []string
	HashParent   [32]byte
}

// NewNode given slot number and transactions
func NewNode(slot uint64, transList []string, keys aesrsa.RSAKeyPair, parent *Node) *Node {
	return &Node{
		Seed:       genesis.Seed,
		Slot:       slot,
		Peer:       aesrsa.KeyToString(keys.Public),
		Draw:       getDraw(slot, genesis.Seed, keys.Private),
		TransList:  transList,
		HashParent: hashNode(parent)}
}

func getDraw(slot, seed uint64, sk aesrsa.RSAKey) []byte {
	json1, err := json.Marshal(slot)
	check(err)
	json2, err := json.Marshal(seed)
	check(err)

	return aesrsa.SignRSA(append(json1, json2...), sk)
}

func hashNode(node *Node) [32]byte {
	json, err := json.Marshal(node)
	check(err)

	return sha256.Sum256(json)
}

func valueOfDraw(node *Node) *big.Int {
	var val big.Int

	json1, err := json.Marshal(node.Slot)
	check(err)
	json2, err := json.Marshal(node.Seed)
	check(err)
	json3, err := json.Marshal(node.Peer)
	check(err)
	json4, err := json.Marshal(node.Peer)
	check(err)
	json := append(json1, json2...)
	json = append(json, json3...)
	json = append(json, json4...)

	hash := sha256.Sum256(json)

	hashInt := new(big.Int).SetBytes(hash[:])

	return val.Mul(hashInt, big.NewInt(getStake(node.Peer)))
}

func getStake(peer string) int64 {
	return int64(ledger.GetBalance(peer))
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
