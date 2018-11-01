package account

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"../aesrsa"
)

// SignedTransaction is an atomic operation on a ledger
type SignedTransaction struct {
	ID        string
	From      string
	To        string
	Amount    int
	Signature string
}

// NewSignedTransaction is a constructor of transactions
func NewSignedTransaction(ID, From, To string, Amount int) *SignedTransaction {
	return &SignedTransaction{
		ID:     ID,
		From:   From,
		To:     To,
		Amount: Amount}
}

// ExtractTransaction extracts the transaction from the signed one
func (st SignedTransaction) ExtractTransaction() Transaction {
	return Transaction{
		ID:     st.ID,
		From:   st.From,
		To:     st.To,
		Amount: st.Amount}
}

// SignTransaction signs a transaction as the sender
func SignTransaction(t Transaction, privKey aesrsa.RSAKey) SignedTransaction {
	jsonT, err := json.Marshal(t)
	check(err)

	sign := base64.StdEncoding.EncodeToString(aesrsa.SignRSA(jsonT, privKey))

	return SignedTransaction{
		ID:        t.ID,
		From:      t.From,
		To:        t.To,
		Amount:    t.Amount,
		Signature: sign}
}

// VerifyTransaction verifies that a transaction signature corresponds to the sender
func (st SignedTransaction) VerifyTransaction() bool {
	t := st.ExtractTransaction()
	jsonT, err := json.Marshal(t)
	check(err)

	sign, err := base64.StdEncoding.DecodeString(st.Signature)
	check(err)

	return aesrsa.VerifyRSA(jsonT, sign, aesrsa.KeyFromString(st.From))
}

// WhatType returns "Block" for SignedTransaction type
func (st SignedTransaction) WhatType() string {
	return "SignedTransaction"
}

func (st SignedTransaction) String() string {
	return fmt.Sprintf("Transaction: ID %s, From %s To %s, Amount %d, Signature %s", st.ID, st.From, st.To, st.Amount, st.Signature)
}
