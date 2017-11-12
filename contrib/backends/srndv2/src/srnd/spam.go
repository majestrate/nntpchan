package srnd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os/user"
	"strings"
)

var ErrSpamFilterNotEnabled = errors.New("spam filter access attempted when disabled")
var ErrSpamFilterFailed = errors.New("spam filter failed")
var ErrMessageIsSpam = errors.New("message is spam")

type SpamFilter struct {
	addr    string
	enabled bool
}

func (sp *SpamFilter) Configure(c SpamConfig) {
	sp.enabled = c.enabled
	sp.addr = c.addr
}

func (sp *SpamFilter) Enabled(newsgroup string) bool {
	return sp.enabled && newsgroup != "ctl"
}

type SpamResult struct {
	Err    error
	IsSpam bool
}

func (sp *SpamFilter) Rewrite(msg io.Reader, out io.WriteCloser, group string) (result SpamResult) {
	var buff [65636]byte
	if !sp.Enabled(group) {
		result.Err = ErrSpamFilterNotEnabled
		return
	}
	var addr *net.TCPAddr
	var c *net.TCPConn
	var u *user.User
	addr, result.Err = net.ResolveTCPAddr("tcp", sp.addr)
	if result.Err != nil {
		return
	}
	c, result.Err = net.DialTCP("tcp", nil, addr)
	if result.Err != nil {
		return
	}
	u, result.Err = user.Current()
	if result.Err != nil {
		return
	}
	fmt.Fprintf(c, "PROCESS SPAMC/1.5\r\nUser: %s\r\n\r\n", u.Username)
	io.CopyBuffer(c, msg, buff[:])
	c.CloseWrite()
	r := bufio.NewReader(c)
	for {
		var l string
		l, result.Err = r.ReadString(10)
		if result.Err != nil {
			return
		}
		l = strings.TrimSpace(l)
		if strings.HasPrefix(l, "Spam: True ") {
			result.IsSpam = true
		}
		log.Println("SpamFilter:", l)
		if l == "" {
			_, result.Err = io.CopyBuffer(out, r, buff[:])
			c.Close()
			out.Close()
			return
		}
	}
	result.Err = ErrSpamFilterFailed
	return
}
