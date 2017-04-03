package tinyhtml

import (
	"io"
	"bufio"
)

//Made to wrap around a file reader to compress html for webserver applications in order to reduce bandwidth
type Minimizer struct {
	inp     *bufio.Reader
	comment bool
	intag   bool
	intext  bool
}

//Creates an html minimizer with the given Reader as its input
func New(i io.Reader) *Minimizer {
	m := new(Minimizer)
	m.inp = bufio.NewReader(i)
	m.intag = false
	m.intext = false
	m.comment = false
	return m
}

//Read compressed html into the buffer given
func (m *Minimizer) Read(b []byte) (int, error) {
	i := 0
	for i < len(b) {
		sb, err := m.inp.ReadByte()
		if err != nil {
			return i, err
		}
		switch sb {
		case '-':
			if m.comment {
				temp := make([]byte, 2)
				_, err := m.inp.Read(temp)
				if err != nil {
					return i, err
				}
				if string(temp) == "->" {
					m.comment = false
				}
				continue
			}
		case '<':
			if m.comment {
				continue
			}
			peek, err := m.inp.Peek(3)
			if err != nil {
				return i, err
			}
			if string(peek) == "!--" {
				m.comment = true
				continue
			}
			m.intag = true
			m.intext = false
		case '>':
			if m.comment {
				continue
			}
			m.intag = false
		case '\n':
			continue
		case '\r':
			continue
		case '\t':
			continue
		case ' ':
			if !m.intext && !m.intag {
				continue
			}
		default:
			if !m.intag {
				m.intext = true
			}
		}
		if !m.comment {
			b[i] = sb
			i++
		}
	}
	return i, nil
}
