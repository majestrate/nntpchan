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

// feed spam subsystem a spam post
func (sp *SpamFilter) MarkSpam(msg io.Reader) (err error) {
	var buf [65636]byte

	var u *user.User
	u, err = user.Current()
	if err != nil {
		return
	}
	var conn *net.TCPConn
	conn, err = sp.openConn()
	if err != nil {
		return
	}
	defer conn.Close()
	fmt.Fprintf(conn, "TELL SPAMC/1.5\r\nUser: %s\r\nMessage-class: spam\r\nSet: local,remote\r\n\r\n", u.Username)
	io.CopyBuffer(conn, msg, buf[:])
	conn.CloseWrite()
	r := bufio.NewReader(conn)
	io.Copy(Discard, r)
	return
}

func (sp *SpamFilter) openConn() (*net.TCPConn, error) {
	addr, err := net.ResolveTCPAddr("tcp", sp.addr)
	if err != nil {
		return nil, err
	}
	return net.DialTCP("tcp", nil, addr)
}

func (sp *SpamFilter) Rewrite(msg io.Reader, out io.WriteCloser, group string) (result SpamResult) {
	var buff [65636]byte
	if !sp.Enabled(group) {
		result.Err = ErrSpamFilterNotEnabled
		return
	}
	var u *user.User
	var c *net.TCPConn
	c, result.Err = sp.openConn()
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
