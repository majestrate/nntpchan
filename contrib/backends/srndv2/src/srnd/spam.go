package srnd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"os/user"
)

var ErrSpamFilterNotEnabled = errors.New("spam filter access attempted when disabled")

type SpamFilter struct {
	addr    string
	enabled bool
}

func (sp *SpamFilter) Configure(c SpamConfig) {
	sp.enabled = c.enabled
	sp.addr = c.addr
}

func (sp *SpamFilter) Enabled() bool {
	return sp.enabled
}

func (sp *SpamFilter) Rewrite(msg io.Reader, out io.WriteCloser) error {
	var buff [65636]byte
	if !sp.Enabled() {
		return ErrSpamFilterNotEnabled
	}
	addr, err := net.ResolveTCPAddr("tcp", sp.addr)
	if err != nil {
		return err
	}
	c, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		return err
	}
	u, err := user.Current()
	if err != nil {
		return err
	}
	fmt.Fprintf(c, "PROCESS SPAMC/1.5\r\nUser: %s\r\n\r\n", u.Username)
	io.CopyBuffer(c, msg, buff[:])
	c.CloseWrite()
	r := bufio.NewReader(c)
	for {
		l, err := r.ReadString(10)
		if err != nil {
			return err
		}
		if len(l) == 1 {
			break
		}
	}
	_, err = io.CopyBuffer(out, r, buff[:])
	c.Close()
	out.Close()
	return err
}
