package account

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"../aesrsa"
)

// SignedBlock is an atomic operation on a ledger
type SignedBlock struct {
	Number    int
	TransList []string
	Signature string
}

func NewSignedBlock(Number int, TransList []string, privKey aesrsa.RSAKey) *SignedBlock {
	b := NewBlock(Number, TransList)
	return SignBlock(*b, privKey)
}

func (sb SignedBlock) ExtractBlock() Block {
	return Block{
		Number:    sb.Number,
		TransList: sb.TransList}
}

func SignBlock(b Block, privKey aesrsa.RSAKey) *SignedBlock {
	jsonT, err := json.Marshal(b)
	check(err)

	sign := base64.StdEncoding.EncodeToString(aesrsa.SignRSA(jsonT, privKey))

	return &SignedBlock{
		Number:    b.Number,
		TransList: b.TransList,
		Signature: sign}
}

func (sb SignedBlock) VerifyBlock(pubKey aesrsa.RSAKey) bool {
	b := sb.ExtractBlock()
	jsonT, err := json.Marshal(b)
	check(err)

	sign, err := base64.StdEncoding.DecodeString(sb.Signature)
	check(err)

	return aesrsa.VerifyRSA(jsonT, sign, pubKey)
}

func (sb SignedBlock) WhatType() string {
	return "SignedBlock"
}

func (sb SignedBlock) String() string {
	return fmt.Sprintf("Block: Number %d", sb.Number)
}
