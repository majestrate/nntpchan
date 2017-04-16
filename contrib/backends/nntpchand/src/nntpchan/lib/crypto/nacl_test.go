package crypto

import (
	"bytes"
	"crypto/rand"
	"io"
	"testing"
)

func TestNaclToPublic(t *testing.T) {
	pk, sk := GenKeypair()
	t_pk := ToPublic(sk)
	if !bytes.Equal(pk, t_pk) {
		t.Logf("%q != %q", pk, t_pk)
		t.Fail()
	}
}

func TestNaclSignVerify(t *testing.T) {
	var msg [1024]byte
	pk, sk := GenKeypair()
	io.ReadFull(rand.Reader, msg[:])

	signer := CreateSigner(sk)
	signer.Write(msg[:])
	sig := signer.Sign()

	verifier := CreateVerifier(pk)
	verifier.Write(msg[:])
	if !verifier.Verify(sig) {
		t.Logf("%q is invalid signature and is %dB long", sig, len(sig))
		t.Fail()
	}
}
