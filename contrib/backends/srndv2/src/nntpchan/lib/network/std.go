package network

import (
	"net"
)

type stdDialer struct {
}

func (sd *stdDialer) Dial(addr string) (c net.Conn, err error) {
	return net.Dial("tcp", addr)
}

var StdDialer = &stdDialer{}
