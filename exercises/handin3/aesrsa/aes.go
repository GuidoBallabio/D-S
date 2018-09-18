package aesrsa

import (
	"crypto/aes"
	"io/ioutil"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

// EncryptToFile to a given file an input string given an AES-key
func EncryptToFile(fin, fout, pw string) {
	pt, err := ioutil.ReadFile(fin)
	check(err)

	cipher, err := aes.NewCipher([]byte(pw))
	check(err)

	ct := make([]byte, len(pt))

	cipher.Encrypt(ct, pt)

	err = ioutil.WriteFile(fout, ct, 0644)
	check(err)
}

// DecryptToFile to a given file an input string given an AES-key
func DecryptToFile(fin, fout, pw string) {
	ct, err := ioutil.ReadFile(fin)
	check(err)

	cipher, err := aes.NewCipher([]byte(pw))
	check(err)

	pt := make([]byte, len(ct))

	cipher.Encrypt(pt, ct)

	err = ioutil.WriteFile(fout, pt, 0644)
	check(err)
}
