package aesrsa

import (
	"crypto/rand"
	"math/big"
)

// RSAKeyPair is a pair of public/private key pair for RSA encryption
type RSAKeyPair struct {
	Public  RSAKey
	Private RSAKey
}

// RSAKey is a key for RSA encryption
type RSAKey struct {
	N, Exp *big.Int
}

var zero = big.NewInt(0)
var one = big.NewInt(1)
var two = big.NewInt(2)

var publicExponent = big.NewInt(3)

// KeyGen generates a key pair for RSA
func KeyGen(bits int) (*RSAKeyPair, error) {
	var n, d, e, pl, ql, phi = big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0)

	e.Set(publicExponent)

	p, err := findPrimeNotCoprime(bits/2, e)
	if err != nil {
		return nil, err
	}

	q, err := findPrimeNotCoprime(bits/2, e)
	if err != nil {
		return nil, err
	}

	for q.Cmp(p) == 0 {
		q, err = findPrimeNotCoprime(bits/2, e)
		if err != nil {
			return nil, err
		}
	}

	n.Mul(p, q)

	pl.Sub(p, one)
	ql.Sub(q, one)

	phi.Mul(pl, ql)

	d.ModInverse(e, phi)

	return &RSAKeyPair{
		Public: RSAKey{
			N:   n,
			Exp: e},
		Private: RSAKey{
			N:   n,
			Exp: d}}, nil
}

func findPrimeNotCoprime(bits int, e *big.Int) (*big.Int, error) {
	var pl, temp big.Int
	temp.Set(one)

	p, err := rand.Prime(rand.Reader, bits)
	if err != nil {
		return nil, err
	}

	for temp.GCD(nil, nil, pl.Sub(p, one), e).Cmp(one) != 0 {
		p, err = rand.Prime(rand.Reader, bits/2)
		if err != nil {
			return nil, err
		}
	}

	return p, nil
}

// Encrypt plaintext big.Int using RSAKey
func Encrypt(pt *big.Int, pubKey RSAKey) *big.Int {
	var ct big.Int

	return ct.Exp(pt, pubKey.Exp, pubKey.N)
}

// Decrypt chipertext big.Int using RSAKey
func Decrypt(ct *big.Int, privKey RSAKey) *big.Int {
	var pt big.Int

	return pt.Exp(ct, privKey.Exp, privKey.N)
}