package aesrsa

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestEncryptDecryptFile(t *testing.T) {
	file := "ciphertext"
	pw := "password"
	ptKeys, err := KeyGen(1024)
	checkTest(err, t)

	pt, err := json.Marshal(ptKeys.Private)
	checkTest(err, t)

	EncryptToFile(pt, file, pw)
	out := DecryptFromFile(file, pw)

	if !bytes.Equal(out, pt) {
		t.Errorf("Plaintext not equal to decrypted ciphertext in AES (bytes)")
	}

	res := RSAKey{}
	err = json.Unmarshal(out, &res)
	checkTest(err, t)

	if res.N.Cmp(ptKeys.Private.N) != 0 && res.Exp.Cmp(ptKeys.Private.Exp) != 0 {
		t.Errorf("Key from file not equal to original one")
	}
}
