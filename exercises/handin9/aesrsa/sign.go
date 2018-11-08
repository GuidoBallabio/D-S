package aesrsa

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"io/ioutil"

	"golang.org/x/crypto/pbkdf2"
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

var saltSize = 64

// Generate wallet (file with private key encrypted) and return public key
func Generate(filename string, password string) string {
	ptKeys, err := KeyGen(2048)
	check(err)

	pt := []byte(KeyToString(ptKeys.Private))

	salt := make([]byte, saltSize)
	_, err = rand.Read(salt)
	check(err)

	keyAes := pbkdf2.Key([]byte(password), salt, 4096, 32, sha256.New)

	ct := encryptAES(pt, keyAes)

	err = ioutil.WriteFile(filename, append(salt, ct...), 0644)
	check(err)

	return KeyToString(ptKeys.Public)
}

// Sign signs a message given a wallet with a private key and relative password
func Sign(filename string, password string, msg []byte) []byte {
	ct, err := ioutil.ReadFile(filename)
	check(err)

	salt := ct[:saltSize]
	ct = ct[saltSize:]

	keyAes := pbkdf2.Key([]byte(password), salt, 4096, 32, sha256.New)

	privKey := KeyFromString(string(decryptAES(ct, keyAes)))

	return SignRSA(msg, privKey)
}
