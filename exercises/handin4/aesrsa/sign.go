package aesrsa

import (
	"bytes"
	"crypto/sha256"
	"math/big"
)

// SignRSA sign an input given a RSA key (private)
func SignRSA(input []byte, privkey RSAKey) []byte {
	// Create hash of the message
	hash := sha256.Sum256(input)

	ptHash := new(big.Int).SetBytes(hash[:])
	return Encrypt(ptHash, privkey).Bytes()
}

// VerifyRSA verify and validate an input given a RSA key (public)
func VerifyRSA(input []byte, sign []byte, pubkey RSAKey) bool {
	// Create hash of the message
	hash := sha256.Sum256(input)

	// Extract signed hash
	ctHash := new(big.Int).SetBytes(sign)
	hashSigned := Decrypt(ctHash, pubkey).Bytes()

	return bytes.Equal(hashSigned, hash[:])
}
