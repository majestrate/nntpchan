package network

import (
	"errors"
	log "github.com/Sirupsen/logrus"
	"io"
	"net"
	"strconv"
	"strings"
)

type socksDialer struct {
	socksAddr string
	dialer    Dialer
}

// try dialing out via socks proxy
func (sd *socksDialer) Dial(remote string) (c net.Conn, err error) {
	log.WithFields(log.Fields{
		"addr":  remote,
		"socks": sd.socksAddr,
	}).Debug("dailing out to socks proxy")
	c, err = sd.dialer.Dial(sd.socksAddr)
	if err == nil {
		// dailed out to socks proxy good
		remote_addr := remote
		// generate request
		idx := strings.LastIndex(remote_addr, ":")
		if idx == -1 {
			err = errors.New("invalid address: " + remote_addr)
			return
		}
		var port uint64
		addr := remote_addr[:idx]
		port, err = strconv.ParseUint(remote_addr[idx+1:], 10, 16)
		if port >= 25536 {
			err = errors.New("bad proxy port")
			c.Close()
			c = nil
			return
		} else if err != nil {
			c.Close()
			return
		}
		var proxy_port uint16
		proxy_port = uint16(port)
		proxy_ident := "srndproxy"
		req_len := len(addr) + 1 + len(proxy_ident) + 1 + 8

		req := make([]byte, req_len)
		// pack request
		req[0] = '\x04'
		req[1] = '\x01'
		req[2] = byte(proxy_port & 0xff00 >> 8)
		req[3] = byte(proxy_port & 0x00ff)
		req[7] = '\x01'
		idx = 8

		proxy_ident_b := []byte(proxy_ident)
		addr_b := []byte(addr)

		var bi int
		for bi = range proxy_ident_b {
			req[idx] = proxy_ident_b[bi]
			idx += 1
		}
		idx += 1
		for bi = range addr_b {
			req[idx] = addr_b[bi]
			idx += 1
		}
		log.WithFields(log.Fields{
			"addr":  remote,
			"socks": sd.socksAddr,
			"req":   req,
		}).Debug("write socks request")
		n := 0
		n, err = c.Write(req)
		if err == nil && n == len(req) {
			// wrote request okay
			resp := make([]byte, 8)
			_, err = io.ReadFull(c, resp)
			if err == nil {
				// got reply okay
				if resp[1] == '\x5a' {
					// successful socks connection
					log.WithFields(log.Fields{
						"addr":  remote,
						"socks": sd.socksAddr,
					}).Debug("socks proxy connection successful")
				} else {
					// unsucessful socks connect
					log.WithFields(log.Fields{
						"addr":  remote,
						"socks": sd.socksAddr,
						"code":  resp[1],
					}).Warn("connect via socks proxy failed")
					c.Close()
					c = nil
				}
			} else {
				// error reading reply
				log.WithFields(log.Fields{
					"addr":  remote,
					"socks": sd.socksAddr,
				}).Error("failed to read socks response ", err)
				c.Close()
				c = nil
			}
		} else {
			if err == nil {
				err = errors.New("short write")
			}

			// error writing request
			log.WithFields(log.Fields{
				"addr":  remote,
				"socks": sd.socksAddr,
			}).Error("failed to write socks request ", err)
			c.Close()
			c = nil

		}
	} else {
		// dail fail
		log.WithFields(log.Fields{
			"addr":  remote,
			"socks": sd.socksAddr,
		}).Error("Cannot connect to socks proxy ", err)
	}
	return
}

// create a socks dialer that dials out via socks proxy at address
func SocksDialer(addr string) Dialer {
	return &socksDialer{
		socksAddr: addr,
		dialer:    StdDialer,
	}
}
