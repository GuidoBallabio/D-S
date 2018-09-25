package aesrsa

import (
	"math/big"
	"testing"
)

func TestKeyGen(t *testing.T) {
	keys, err := KeyGen(10)
	checkTest(err, t)

	priv, pub := keys.Public, keys.Private

	n := priv.N
	e := pub.Exp
	d := priv.Exp

	if pub.N.Cmp(n) != 0 {
		t.Errorf("n = pq is different for public and private keys: %d, %d", pub.N.Int64(), priv.N.Int64())
	}

	pfs := primeFactors(n)
	if len(pfs) != 2 {
		t.Errorf("n has more than 2 factors")
	}

	prod := prodSlice(pfs)
	if prod.Cmp(n) != 0 {
		t.Errorf("n != pq : %d != %d", n.Int64(), prod.Int64())
	}

	var k, pl, ql, unit = big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0)
	pl.Sub(&pfs[0], one)
	ql.Sub(&pfs[1], one)
	k.Mul(pl, ql)

	unit.Mul(d, e)
	unit.Mod(one, k)
	if unit.Cmp(one) != 0 {
		t.Errorf("d is not inverse of e mod (p-1)(q-1): %d * %d != 1 mod (%d)(%d)", e.Int64(), d.Int64(), pl.Int64(), ql.Int64())
	}
}

func TestFindPrimeNotCoprime(t *testing.T) {
	var gcd, pl big.Int

	for _, v := range []int64{3, 7, 17} {
		p, err := findPrimeNotCoprime(16, big.NewInt(v))
		checkTest(err, t)

		if gcd.GCD(nil, nil, pl.Sub(p, one), big.NewInt(v)).Cmp(one) != 0 {
			t.Errorf("p-1 (%d) and e (%d) are not coprime: GCD (%d)", pl.Int64(), v, gcd.Int64())
		}
	}

}

func TestEncryptDecrypt(t *testing.T) {
	keys, err := KeyGen(10)
	checkTest(err, t)

	pt := big.NewInt(84)

	ct := Encrypt(pt, keys.Public)
	temp := Decrypt(ct, keys.Private)

	if temp.Cmp(pt) != 0 {
		t.Errorf("Plaintext not equal to decrypted ciphertext: %d != %d", pt.Int64(), temp.Int64())
	}
}

func checkTest(err error, t *testing.T) {
	if err != nil {
		t.Errorf(err.Error())
	}
}

func primeFactors(n *big.Int) (pfs []big.Int) {
	var mod, pf, temp big.Int
	temp.Set(n)

	// Get the number of 2s that divide n
	pf.Set(two)
	for mod.Mod(&temp, &pf).Cmp(zero) == 0 {
		pfs = append(pfs, pf)
	}

	// n must be odd at this point. so we can skip one element
	// (note i = i + 2)

	for pf.SetUint64(3); n.Cmp(prodSlice(pfs)) != 0; pf.Add(&pf, two) {
		var pf2 big.Int
		pf2.Set(&pf)

		// while i divides n, append i and divide n
		for mod.Mod(&temp, &pf).Cmp(zero) == 0 {
			pfs = append(pfs, pf2)
			temp.Div(&temp, &pf2)
		}
	}

	return
}

func prodSlice(s []big.Int) *big.Int {
	prod := big.NewInt(1)
	for _, v := range s {
		prod.Mul(prod, &v)
	}
	return prod
}
