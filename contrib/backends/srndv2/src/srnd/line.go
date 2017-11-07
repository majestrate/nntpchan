package srnd

import (
	"bytes"
	"io"
)

type LineWriter struct {
	w     io.Writer
	limit int
}

func NewLineWriter(w io.Writer, limit int) *LineWriter {
	return &LineWriter{
		w:     w,
		limit: limit,
	}
}

func (l *LineWriter) Write(data []byte) (n int, err error) {
	dl := len(data)
	data = bytes.Replace(data, []byte{13, 10}, []byte{10}, -1)
	parts := bytes.Split(data, []byte{10})
	for _, part := range parts {
		for len(part) > l.limit {
			d := make([]byte, l.limit)
			copy(d, part[:l.limit])
			d = append(d, 10)
			_, err = l.w.Write(d)
			part = part[l.limit:]
			if err != nil {
				return
			}
		}
		part = append(part, 10)
		_, err = l.w.Write(part)
	}
	n = dl
	return
}
