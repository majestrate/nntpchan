package util

import (
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"nntpchan/lib/crypto/nacl"
)

// generate a login salt for nntp users
func GenLoginCredSalt() (salt string) {
	salt = randStr(128)
	return
}

// do nntp login credential hash given password and salt
func NntpLoginCredHash(passwd, salt string) (str string) {
	var b []byte
	b = append(b, []byte(passwd)...)
	b = append(b, []byte(salt)...)
	h := sha512.Sum512(b)
	str = base64.StdEncoding.EncodeToString(h[:])
	return
}

// make a random string
func randStr(length int) string {
	return hex.EncodeToString(nacl.RandBytes(length))[length:]
}
