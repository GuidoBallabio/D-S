package aesrsa

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
)

// SignRSA sign an input given a RSA key
func SignRSA(input []byte, key RSAKey) []byte {
	rand.Int
	sha256.BlockSize
	bytes.Equal
}

// VerifyRSA verify and validate an input given a RSA key
func VerifyRSA(input []byte, sign []byte, key RSAKey) {

}
