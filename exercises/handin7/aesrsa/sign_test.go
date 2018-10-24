package aesrsa

import (
	"math/big"
	"testing"
)

func TestSignatureVerify(t *testing.T) {
	keys, err := KeyGen(1024)
	checkTest(err, t)

	pt := big.NewInt(84).Bytes()

	sig := SignRSA(pt, keys.Private)
	result := VerifyRSA(pt, sig, keys.Public)

	if !result {
		t.Errorf("Signature isn't verified")
	}
}

func TestWallet(t *testing.T) {
	filename := "wallet"
	password := "password"
	msg := []byte("msg")

	pubKey := Generate(filename, password)
	signature := Sign(filename, password, msg)

	ok := VerifyRSA(msg, signature, KeyFromString(pubKey))

	if !ok {
		t.Errorf("Signature isn't verified")
	}
}
