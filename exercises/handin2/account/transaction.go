package account

import (
	"fmt"
)

// Transaction is an atomic operation on a ledger
type Transaction struct {
	ID     string
	From   string
	To     string
	Amount int
}

// NewTransaction is a constructor of transactions
func NewTransaction(ID, From, To string, Amount int) *Transaction {
	return &Transaction{
		ID:     ID,
		From:   From,
		To:     To,
		Amount: Amount}
}

func (t Transaction) String() string {
	return fmt.Sprintf("Transaction: ID\t%s, From\t%s To\t%s, Amount\t%d", t.ID, t.From, t.To, t.Amount)
}
