package aesrsa

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math/big"
)

// SignRSA sign an input given a RSA key (private)
func SignRSA(input []byte, privkey RSAKey) []byte {
	// Create hash of the message
	hash := sha256.Sum256(input)

	ptHash := new(big.Int).SetBytes(hash[:])
	fmt.Println(input, hash, ptHash)
	return Encrypt(ptHash, privkey).Bytes()
}

// VerifyRSA verify and validate an input given a RSA key (public)
func VerifyRSA(input []byte, sign []byte, pubkey RSAKey) bool {
	// Create hash of the message
	hash := sha256.Sum256(input)

	ctHash := new(big.Int).SetBytes(sign)
	// Recover the hash from the signature
	hashSigned := Decrypt(ctHash, pubkey).Bytes()

	fmt.Println(input, hash, sign, ctHash, hashSigned)

	return bytes.Equal(hashSigned, hash[:])
}
