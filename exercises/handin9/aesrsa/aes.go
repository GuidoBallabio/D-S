package aesrsa

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"errors"
	"io/ioutil"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

// encryptAES returns the ciphertext  of the plain text given the key in bytes
func encryptAES(pt, key []byte) []byte {
	// creating block
	block, err := aes.NewCipher(key)
	check(err)

	// padding plainttext to match blocksize
	pt, err = pkcs7Pad(pt, aes.BlockSize)
	check(err)

	// allocating slice for ciphertext plus iv (len(iv) == aes.Blocksize)
	ct := make([]byte, len(pt)+aes.BlockSize)

	// using first block of allocated data for cipher text as random initializator
	iv := ct[:aes.BlockSize]
	_, err = rand.Read(iv)
	check(err)

	// creating cipher from block and IV
	streamCipher := cipher.NewCTR(block, iv)

	// encrypting message but not iv
	streamCipher.XORKeyStream(ct[aes.BlockSize:], pt)

	return ct
}

// decryptAES returns the ciphertext  of the plain text given the key in bytes
func decryptAES(ct, key []byte) []byte {
	// // padding key
	// pwBytes, err := pkcs7Pad(key, 32)
	// check(err)

	// creating block
	block, err := aes.NewCipher(key)
	check(err)

	// dividing IV and proper ct
	iv := ct[:aes.BlockSize]
	ct = ct[aes.BlockSize:]

	if (len(ct) % aes.BlockSize) != 0 {
		panic(errors.New("Ciphertext should have been multiple of aes.Blocksize"))
	}

	// allocating for plaintext without IV
	pt := make([]byte, len(ct))

	streamCipher := cipher.NewCTR(block, iv)
	streamCipher.XORKeyStream(pt, ct)

	// unpad plaintext
	pt, err = pkcs7Unpad(pt, aes.BlockSize)
	check(err)

	return pt
}

// EncryptToFile to a given file an input string given an AES-key
func EncryptToFile(pt []byte, fout, pw string) {
	// padding key to make aes-256
	pwBytes, err := pkcs7Pad([]byte(pw), 32)
	check(err)

	ct := encryptAES(pt, pwBytes)
	err = ioutil.WriteFile(fout, ct, 0644)
	check(err)
}

// DecryptFromFile to a given file an input string given an AES-key
func DecryptFromFile(fin, pw string) []byte {
	ct, err := ioutil.ReadFile(fin)
	check(err)

	// padding key to make aes-256
	pwBytes, err := pkcs7Pad([]byte(pw), 32)
	check(err)

	return decryptAES(ct, pwBytes)

}

// StoreKeyPair writes a RSAKeyPair to a file (encrypted)
func StoreKeyPair(keys *RSAKeyPair, file, pw string) {
	pt, err := json.Marshal(keys)
	check(err)

	EncryptToFile(pt, file, pw)
}

// ReadKeyPair retrieves a RSAKeyPair from a file (decrypting it)
func ReadKeyPair(file, pw string) *RSAKeyPair {
	out := DecryptFromFile(file, pw)

	res := &RSAKeyPair{}
	err := json.Unmarshal(out, &res)
	check(err)

	return res
}

// This two function below have been copied from https://github.com/go-web/tokenizer/blob/master/pkcs7.go
// as is not a good idea to reinvent padding, especially in a cryptographic context

// PKCS7 errors.
var (
	// ErrInvalidBlockSize indicates hash blocksize <= 0.
	ErrInvalidBlockSize = errors.New("invalid blocksize")

	// ErrInvalidPKCS7Data indicates bad input to PKCS7 pad or unpad.
	ErrInvalidPKCS7Data = errors.New("invalid PKCS7 data (empty or not padded)")

	// ErrInvalidPKCS7Padding indicates PKCS7 unpad fails to bad input.
	ErrInvalidPKCS7Padding = errors.New("invalid padding on input")
)

// pkcs7Pad right-pads the given byte slice with 1 to n bytes, where
// n is the block size. The size of the result is x times n, where x
// is at least 1.
func pkcs7Pad(b []byte, blocksize int) ([]byte, error) {
	if blocksize <= 0 {
		return nil, ErrInvalidBlockSize
	}
	if b == nil || len(b) == 0 {
		return nil, ErrInvalidPKCS7Data
	}
	n := blocksize - (len(b) % blocksize)
	pb := make([]byte, len(b)+n)
	copy(pb, b)
	copy(pb[len(b):], bytes.Repeat([]byte{byte(n)}, n))
	return pb, nil
}

// pkcs7Unpad validates and unpads data from the given bytes slice.
// The returned value will be 1 to n bytes smaller depending on the
// amount of padding, where n is the block size.
func pkcs7Unpad(b []byte, blocksize int) ([]byte, error) {
	if blocksize <= 0 {
		return nil, ErrInvalidBlockSize
	}
	if b == nil || len(b) == 0 {
		return nil, ErrInvalidPKCS7Data
	}
	if len(b)%blocksize != 0 {
		return nil, ErrInvalidPKCS7Padding
	}
	c := b[len(b)-1]
	n := int(c)
	if n == 0 || n > len(b) {
		return nil, ErrInvalidPKCS7Padding
	}
	for i := 0; i < n; i++ {
		if b[len(b)-n+i] != c {
			return nil, ErrInvalidPKCS7Padding
		}
	}
	return b[:len(b)-n], nil
}
