package services

import (
	"fmt"
	"time"

	. "../account"
	"../aesrsa"
)

var sequencer aesrsa.RSAKey
var sequencerSecret aesrsa.RSAKey

// BecomeSequencer initialize the sequencer
func BecomeSequencer() {
	keyPair, err := aesrsa.KeyGen(2048)

	if err != nil {
		fmt.Println(err.Error())
	}

	sequencer = keyPair.Public
	sequencerSecret = keyPair.Private
}

// CheckIfSequencer returns true if the local machine is the sequencer
func CheckIfSequencer() bool {
	return sequencerSecret != aesrsa.RSAKey{}
}

// BeSequencer add the beheaviour of a sequencer to the peer
func BeSequencer(sequencerCh <-chan Transaction, blockCh chan<- SignedBlock, quitCh <-chan struct{}) {
	defer Wg.Done()

	fmt.Println("You are the Sequencer")

	var n int
	ticker := time.NewTicker(time.Second * 10)

	for {
		seq := make([]string, 0)
		endBlock := false
		for !endBlock {
			select {
			case <-ticker.C:
				if len(seq[:]) > 0 {
					sb := NewSignedBlock(n, seq, sequencerSecret)
					broadcastBlock(*sb)
					blockCh <- *sb
					n++
					endBlock = true
				}
			case t := <-sequencerCh:
				seq = append(seq, t.ID)
			case <-quitCh:
				return //Done
			}
		}
	}
}
