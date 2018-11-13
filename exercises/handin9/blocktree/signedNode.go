package blocktree

import (
	"encoding/base64"
	"encoding/json"

	"../aesrsa"
)

// A SignedNode of the Tree
type SignedNode struct {
	Node
	Signature string
}

// NewSignedNode creates a SignedNode from a node
func NewSignedNode(node Node, sk aesrsa.RSAKey) *SignedNode {
	jsonT, err := json.Marshal(node)
	check(err)

	sign := base64.StdEncoding.EncodeToString(aesrsa.SignRSA(jsonT, sk))

	return &SignedNode{
		Node:      node,
		Signature: sign}
}

// VerifyNode verifies that a node signature corresponds to the sender
func (sn SignedNode) VerifyNode() bool {
	n := sn.Node
	jsonT, err := json.Marshal(n)
	check(err)

	sign, err := base64.StdEncoding.DecodeString(sn.Signature)
	check(err)

	return aesrsa.VerifyRSA(jsonT, sign, aesrsa.KeyFromString(n.Peer))
}

// WhatType returns "SignedNode" for SignedNode type
func (sn SignedNode) WhatType() string {
	return "SignedNode"
}
