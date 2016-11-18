package crypto

import "io"

// a detached signature
type Signature []byte

type SigEncoder interface {
	// encode a signature to an io.Writer
	// return error if one occurrened while writing out signature
	Encode(sig Signature, w io.Writer) error
	// encode a signature to a string
	EncodeString(sig Signature) string
}

// a decoder of signatures
type SigDecoder interface {
	// decode signature from io.Reader
	// reads all data until io.EOF
	// returns singaure or error if an error occured while reading
	Decode(r io.Reader) (Signature, error)
	// decode a signature from string
	// returns signature or error if an error ocurred while decoding
	DecodeString(str string) (Signature, error)
}
