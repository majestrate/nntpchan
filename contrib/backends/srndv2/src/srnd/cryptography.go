// +build !libsodium

package srnd

import (
	"crypto/sha512"
	"golang.org/x/crypto/curve25519"
	"golang.org/x/crypto/ed25519"
)

func naclCryptoVerifyFucky(h, sig, pk []byte) bool {
	pub := make(ed25519.PublicKey, ed25519.PublicKeySize)
	copy(pub, pk)
	return ed25519.Verify(pub, h, sig)
}

func naclCryptoSignFucky(hash, sk []byte) []byte {
	sec := make(ed25519.PrivateKey, ed25519.PrivateKeySize)
	copy(sec, sk)
	return ed25519.Sign(sec, hash)
}

func naclSeedToKeyPair(seed []byte) (pk, sk []byte) {
	h := sha512.Sum512(seed)
	sk = h[:]
	sk[0] &= 248
	sk[31] &= 127
	sk[31] |= 64

	// scalarmult magick shit
	pk = scalarBaseMult(sk)

	copy(sk[0:32], seed[0:32])
	copy(sk[32:64], pk[0:32])

	return
}

func scalarBaseMult(sk []byte) (pk []byte) {
	var skey [32]byte
	var pkey [32]byte
	copy(skey[0:32], sk[0:32])
	curve25519.ScalarBaseMult(&pkey, &skey)
	pk = pkey[0:32]
	return
}
