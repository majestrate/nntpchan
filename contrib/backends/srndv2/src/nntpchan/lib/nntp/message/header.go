package message

import (
	"io"
	"mime"
	"strings"
)

// an nntp message header
type Header map[string][]string

// get message-id header
func (self Header) MessageID() (v string) {
	for _, hdr := range []string{"MessageID", "Message-ID", "Message-Id", "message-id"} {
		v = self.Get(hdr, "")
		if v != "" {
			break
		}
	}
	return
}

func (self Header) Reference() (ref string) {
	return self.Get("Reference", self.MessageID())
}

// extract media type from content-type header
func (self Header) GetMediaType() (mediatype string, params map[string]string, err error) {
	return mime.ParseMediaType(self.Get("Content-Type", "text/plain"))
}

// is this header for a multipart message?
func (self Header) IsMultipart() bool {
	return strings.HasPrefix(self.Get("Content-Type", "text/plain"), "multipart/mixed")
}

func (self Header) IsSigned() bool {
	return self.Get("X-Pubkey-Ed25519", "") != ""
}

func (self Header) Newsgroup() string {
	return self.Get("Newsgroups", "overchan.discard")
}

// do we have a key in this header?
func (self Header) Has(key string) bool {
	_, ok := self[key]
	return ok
}

// set key value
func (self Header) Set(key, val string) {
	self[key] = []string{val}
}

func (self Header) AppendPath(name string) {
	p := self.Get("Path", name)
	if p != name {
		p = name + "!" + p
	}
	self.Set("Path", p)
}

// append value to key
func (self Header) Add(key, val string) {
	if self.Has(key) {
		self[key] = append(self[key], val)
	} else {
		self.Set(key, val)
	}
}

// get via key or return fallback value
func (self Header) Get(key, fallback string) string {
	val, ok := self[key]
	if ok {
		str := ""
		for _, k := range val {
			str += k + ", "
		}
		return str[:len(str)-2]
	} else {
		return fallback
	}
}

// interface for types that can read an nntp header
type HeaderReader interface {
	// blocking read an nntp header from an io.Reader
	// return the read header and nil on success
	// return nil and an error if an error occurred while reading
	ReadHeader(r io.Reader) (Header, error)
}

// interface for types that can write an nntp header
type HeaderWriter interface {
	// blocking write an nntp header to an io.Writer
	// returns an error if one occurs otherwise nil
	WriteHeader(hdr Header, w io.Writer) error
}

// implements HeaderReader and HeaderWriter
type HeaderIO struct {
	delim byte
}

// read header
func (s *HeaderIO) ReadHeader(r io.Reader) (hdr Header, err error) {
	hdr = make(Header)
	var k, v string
	var buf [1]byte
	for err == nil {
		// read key
		for err == nil {
			_, err = r.Read(buf[:])
			if err != nil {
				return
			}
			if buf[0] == 58 { // colin
				// consume space
				_, err = r.Read(buf[:])
				for err == nil {
					_, err = r.Read(buf[:])
					if buf[0] == s.delim {
						// got delimiter
						hdr.Add(k, v)
						k = ""
						v = ""
						break
					} else {
						v += string(buf[:])
					}
				}
				break
			} else if buf[0] == s.delim {
				// done
				return
			} else {
				k += string(buf[:])
			}
		}
	}
	return
}

// write header
func (s *HeaderIO) WriteHeader(hdr Header, wr io.Writer) (err error) {
	for k, vs := range hdr {
		for _, v := range vs {
			var line []byte
			// key
			line = append(line, []byte(k)...)
			// ": "
			line = append(line, 58, 32)
			// value
			line = append(line, []byte(v)...)
			// delimiter
			line = append(line, s.delim)
			// write line
			_, err = wr.Write(line)
			if err != nil {
				return
			}
		}
	}
	_, err = wr.Write([]byte{s.delim})
	return
}

func NewHeaderIO() *HeaderIO {
	return &HeaderIO{
		delim: 10,
	}
}
