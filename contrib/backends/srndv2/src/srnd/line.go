package srnd

import (
	"bytes"
	"io"
)

type LineWriter struct {
	w    io.Writer
	Left int64
}

func NewLineWriter(w io.Writer, limit int64) *LineWriter {
	return &LineWriter{
		w:    w,
		Left: limit,
	}
}

func (l *LineWriter) Write(data []byte) (n int, err error) {
	if l.Left <= 0 {
		err = ErrOversizedMessage
		return
	}
	wr := len(data)
	data = bytes.Replace(data, []byte{13, 10}, []byte{10}, -1)
	n, err = l.w.Write(data)
	l.Left -= int64(n)
	if err != nil {
		return n, err
	}
	n = wr
	return
}
