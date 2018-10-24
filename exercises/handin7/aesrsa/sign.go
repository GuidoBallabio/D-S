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

// Generate wallet (file with private key encrypted) and return public key
func Generate(filename string, password string) string {
	ptKeys, err := KeyGen(2048)
	check(err)

	pt := []byte(KeyToString(ptKeys.Private))

	EncryptToFile(pt, filename, password)
	return KeyToString(ptKeys.Public)
}

// Sign signs a message given a wallet with a private key and relative password
func Sign(filename string, password string, msg []byte) []byte {
	privKey := KeyFromString(string(DecryptFromFile(filename, password)))

	return SignRSA(msg, privKey)
}
