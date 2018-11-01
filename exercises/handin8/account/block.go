package account

import (
	"fmt"
)

type Block struct {
	Number int
	TransList []string
}

func NewBlock(Number int, TransList []string){
	return &Block{
		Number:	   Number
		TransList: TransList
	}
}

