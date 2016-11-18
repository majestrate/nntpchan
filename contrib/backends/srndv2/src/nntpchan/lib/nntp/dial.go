package nntp

import (
	"crypto/tls"
	"nntpchan/lib/network"
)

// establishes an outbound nntp connection to a remote server
type Dialer interface {
	// dial out with a dialer
	// if cfg is not nil, try to establish a tls connection with STARTTLS
	// returns a new nntp connection and nil on successful handshake and login
	// returns nil and an error if an error happened
	Dial(d network.Dialer, cfg *tls.Config) (*Conn, error)
}
