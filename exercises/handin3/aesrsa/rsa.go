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
	n, exp *big.Int
}

var publicExponent = big.NewInt(3)

// KeyGen generates a key pair for RSA
func KeyGen(bits int) (*RSAKeyPair, error) {
	var n, k, e, d, pl, ql *big.Int

	e.Set(publicExponent)

	p, err := findPrimeNotCoprime(bits/2, e)
	if err != nil {
		return nil, err
	}

	q, err := findPrimeNotCoprime(bits/2, e)
	if err != nil {
		return nil, err
	}

	n.Mul(p, q)

	pl.Sub(p, big.NewInt(1))
	ql.Sub(q, big.NewInt(1))

	k.Mul(pl, ql)

	d.ModInverse(e, k)

	return &RSAKeyPair{
		Public: RSAKey{
			n:   n,
			exp: e},
		Private: RSAKey{
			n:   n,
			exp: d}}, nil
}

func findPrimeNotCoprime(bits int, e *big.Int) (p *big.Int, err error) {
	var pl, temp *big.Int
	temp.Set(e)

	p, err = rand.Prime(rand.Reader, bits)
	if err != nil {
		return nil, err
	}

	for temp.GCD(nil, nil, pl.Sub(p, big.NewInt(1)), e).Cmp(big.NewInt(1)) == 0 {
		p, err = rand.Prime(rand.Reader, bits/2)
		if err != nil {
			return nil, err
		}
	}

	return p, nil
}

// Encrypt plaintext big.Int using RSAKey
func Encrypt(pt *big.Int, pubKey RSAKey) *big.Int {
	var ct *big.Int

	return ct.Exp(pt, pubKey.exp, pubKey.n)
}

// Decrypt chipertext big.Int using RSAKey
func Decrypt(ct *big.Int, privKey RSAKey) *big.Int {
	var pt *big.Int

	return pt.Exp(pt, privKey.exp, privKey.n)
}
