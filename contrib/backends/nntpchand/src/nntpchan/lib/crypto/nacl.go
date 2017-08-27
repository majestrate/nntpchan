package crypto

import (
	"crypto/sha512"
	"hash"
	"nntpchan/lib/crypto/nacl"
)

type fuckyNacl struct {
	k    []byte
	hash hash.Hash
}

func (fucky *fuckyNacl) Write(d []byte) (int, error) {
	return fucky.hash.Write(d)
}

func (fucky *fuckyNacl) Sign() (s Signature) {
	h := fucky.hash.Sum(nil)
	if h == nil {
		panic("fuck.hash.Sum == nil")
	}

	_, sec := nacl.SeedToKeyPair(fucky.k)
	sig := nacl.CryptoSignFucky(h, sec)
	if sig == nil {
		panic("fucky signer's call to nacl.CryptoSignFucky returned nil")
	}
	s = Signature(sig)
	fucky.resetState()
	return
}

// reset inner state so we can reuse this fuckyNacl for another operation
func (fucky *fuckyNacl) resetState() {
	fucky.hash = sha512.New()
}

func (fucky *fuckyNacl) Verify(sig Signature) (valid bool) {
	h := fucky.hash.Sum(nil)
	if h == nil {
		panic("fucky.hash.Sum == nil")
	}
	valid = nacl.CryptoVerifyFucky(h, sig, fucky.k)
	fucky.resetState()
	return
}

func createFucky(k []byte) *fuckyNacl {
	return &fuckyNacl{
		k:    k,
		hash: sha512.New(),
	}
}

// create a standard signer given a secret key
func CreateSigner(sk []byte) Signer {
	return createFucky(sk)
}

// create a standard verifier given a public key
func CreateVerifier(pk []byte) Verifer {
	return createFucky(pk)
}

// get the public component given the secret key
func ToPublic(sk []byte) (pk []byte) {
	pk, _ = nacl.SeedToKeyPair(sk)
	return
}

// create a standard keypair
func GenKeypair() (pk, sk []byte) {
	sk = RandBytes(32)
	pk, _ = nacl.SeedToKeyPair(sk)
	return
}
