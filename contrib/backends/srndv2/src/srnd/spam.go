package srnd

import (
	"io"
	"net"
)

type SpamFilter struct {
	addr    string
	enabled bool
}

func (sp *SpamFilter) Configure(c SpamConfig) {
	sp.enabled = c.enabled
	sp.addr = c.addr
}

func (sp *SpamFilter) Rewrite(msg io.Reader, out io.WriteCloser) error {
	var buff [65636]byte
	if sp.enabled {
		addr, err := net.ResolveTCPAddr("tcp", sp.addr)
		if err != nil {
			return err
		}
		c, err := net.DialTCP("tcp", nil, addr)
		if err != nil {
			return err
		}
		io.CopyBuffer(c, msg, buff[:])
		c.CloseWrite()
		_, err = io.CopyBuffer(out, c, buff[:])
		c.Close()
		out.Close()
		return err
	}
	io.CopyBuffer(out, msg, buff[:])
	out.Close()
	return nil
}
