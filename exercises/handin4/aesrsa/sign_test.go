package aesrsa

import (
	"math/big"
	"testing"
)

func TestSignatureVerify(t *testing.T) {
	keys, err := KeyGen(100)
	checkTest(err, t)

	pt := big.NewInt(84000).Bytes()

	sig := SignRSA(pt, keys.Private)
	result := VerifyRSA(pt, sig, keys.Public)

	if !result {
		t.Errorf("Signature isn't verified")
	}
}
