package srnd

import (
	"bytes"
	"io"
)

type LineWriter struct {
	w     io.Writer
	limit int64
}

func NewLineWriter(w io.Writer, limit int64) *LineWriter {
	return &LineWriter{
		w:     w,
		limit: limit,
	}
}

func (l *LineWriter) Write(data []byte) (n int, err error) {
	n = len(data)
	data = bytes.Replace(data, []byte{13, 10}, []byte{10}, -1)
	_, err = l.w.Write(data)
	l.limit -= int64(n)
	if l.limit <= 0 {
		err = ErrOversizedMessage
	}
	return
}
