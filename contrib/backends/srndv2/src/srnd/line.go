package srnd

import (
	"bytes"
	"io"
)

type LineWriter struct {
	w io.Writer
}

func NewLineWriter(w io.Writer) *LineWriter {
	return &LineWriter{
		w: w,
	}
}

func (l *LineWriter) Write(data []byte) (n int, err error) {
	n = len(data)
	data = bytes.Replace(data, []byte{13, 10}, []byte{10}, -1)
	_, err = l.w.Write(data)
	return
}
