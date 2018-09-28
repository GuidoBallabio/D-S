package aesrsa

import (
	"bytes"
	"crypto/sha256"
)

// SignRSA sign an input given a RSA key (private)
func SignRSA(input []byte, privkey RSAKey) []byte {
	// Create hash of the message
	hash := sha256.Sum256(input)

	return EncryptBytes(hash[:], privkey)
}

// VerifyRSA verify and validate an input given a RSA key (public)
func VerifyRSA(input []byte, sign []byte, pubkey RSAKey) bool {
	// Create hash of the message
	hash := sha256.Sum256(input)

	// Extract signed hash
	hashSigned := DecryptBytes(sign, pubkey)

	return bytes.Equal(hashSigned, hash[:])
}
