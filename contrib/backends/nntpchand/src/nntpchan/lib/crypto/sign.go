package crypto

import "io"

//
// provides generic signing interface for producing detached signatures
// call Write() to feed data to be signed, call Sign() to generate
// a detached signature
//
type Signer interface {
	io.Writer
	// generate detached Signature from previously fed body via Write()
	Sign() Signature
}
