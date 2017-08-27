package crypto

import (
	"crypto/rand"
	"io"
)

// generate random bytes
func RandBytes(n int) []byte {
	b := make([]byte, n)
	io.ReadFull(rand.Reader, b)
	return b
}
