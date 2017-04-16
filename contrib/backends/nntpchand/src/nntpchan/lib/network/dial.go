package network

import (
	"errors"
	"net"
	"nntpchan/lib/config"
	"strings"
)

// operation timed out
var ErrTimeout = errors.New("timeout")

// the operation was reset abruptly
var ErrReset = errors.New("reset")

// the operation was actively refused
var ErrRefused = errors.New("refused")

// generic dialer
// dials out to a remote address
// returns a net.Conn and nil on success
// returns nil and error if an error happens while dialing
type Dialer interface {
	Dial(remote string) (net.Conn, error)
}

// create a new dialer from configuration
func NewDialer(conf *config.ProxyConfig) (d Dialer) {
	d = StdDialer
	if conf != nil {
		proxyType := strings.ToLower(conf.Type)
		if proxyType == "socks" || proxyType == "socks4a" {
			d = SocksDialer(conf.Addr)
		}
	}
	return
}
