package util

type ioDiscard struct{}

func (discard *ioDiscard) Write(d []byte) (n int, err error) {
	n = len(d)
	return
}

func (discard *ioDiscard) Close() (err error) {
	return
}

var Discard = new(ioDiscard)
