// +build !libsodium

package srnd

import (
	"golang.org/x/crypto/curve25519"
)

func naclCryptoVerifyFucky(h, sig, pk []byte) bool {
	return false
}

func naclCryptoSignFucky(hash, sk []byte) []byte {
	return nil
}

func naclCryptoVerifyDetached(hash, sig, pk []byte) bool {
	return false
}

func naclCryptoSignDetached(hash, sk []byte) []byte {
	return nil
}

func naclSeedToKeyPair(seed []byte) (pk, sk []byte) {
	return
}

var naclScalarBaseMult = curve25519.ScalarBaseMult
