package crypto

import "io"

// provides generic signature
// call Write() to feed in message body
// once the entire body has been fed in via Write() call Verify() with detached
// signature to verify the detached signature against the previously fed body
type Verifer interface {
	io.Writer
	// verify detached signature from body previously fed via Write()
	// return true if the detached signature is valid given the body
	Verify(sig Signature) bool
}
