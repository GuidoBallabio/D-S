package account

// Block struct defines an order of transactions
type Block struct {
	Number    int
	TransList []string
}

// NewBlock creates new Block
func NewBlock(Number int, TransList []string) *Block {
	return &Block{
		Number:    Number,
		TransList: TransList}
}
