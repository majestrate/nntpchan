package nntp

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// defines server side authentication mechanism
type ServerAuth interface {
	// check plaintext login
	// returns nil on success otherwise error if one occurs during authentication
	// returns true if authentication was successful and an error if a network io error happens
	CheckLogin(username, passwd string) (bool, error)
}

type FlatfileAuth string

func (fname FlatfileAuth) CheckLogin(username, passwd string) (found bool, err error) {
	cred := fmt.Sprintf("%s:%s", username, passwd)
	var f *os.File
	f, err = os.Open(string(fname))
	if err == nil {
		defer f.Close()
		r := bufio.NewReader(f)
		for err == nil {
			var line string
			line, err = r.ReadString(10)
			line = strings.Trim(line, "\r\n")
			if line == cred {
				found = true
				break
			}
		}
	}
	return
}
