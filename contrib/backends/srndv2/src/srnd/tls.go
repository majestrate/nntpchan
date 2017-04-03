package srnd

import (
	"bufio"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"math/big"
	"net"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var TlsNotSupported = errors.New("TLS not supported")
var TlsFailedToLoadCA = errors.New("could not load CA files")

// handle STARTTLS on connection
func HandleStartTLS(conn net.Conn, config *tls.Config) (econn *textproto.Conn, state tls.ConnectionState, err error) {
	if config == nil {
		_, err = io.WriteString(conn, "580 can not intitiate TLS negotiation\r\n")
		if err == nil {
			err = TlsNotSupported
		}
	} else {
		_, err = io.WriteString(conn, "382 Continue with TLS negotiation\r\n")
		if err == nil {
			// begin tls crap here
			tconn := tls.Server(conn, config)
			err = tconn.Handshake()
			state = tconn.ConnectionState()
			if err == nil {
				econn = textproto.NewConn(tconn)
				return
			} else {
				certs := state.PeerCertificates
				if len(certs) == 0 {
					log.Println("starttls failed, no peer certs provided")
				} else {
					for _, cert := range certs {
						for _, dns := range cert.DNSNames {
							log.Println("starttls peer cert from", dns, "not valid")
						}
					}
				}
				tconn.Close()
			}
		}
	}
	return
}

func SendStartTLS(conn net.Conn, config *tls.Config) (econn *textproto.Conn, state tls.ConnectionState, err error) {
	_, err = io.WriteString(conn, "STARTTLS\r\n")
	if err == nil {
		r := bufio.NewReader(conn)
		var line string
		line, err = r.ReadString(10)
		if strings.HasPrefix(line, "382 ") {
			// we gud
			tconn := tls.Client(conn, config)
			// tls okay
			log.Println("TLS Handshake done", config.ServerName)
			state = tconn.ConnectionState()
			econn = textproto.NewConn(tconn)
			return
		} else {
			// it won't do tls
			err = TlsNotSupported
		}
		r = nil
	}
	return
}

// create base tls certificate
func newTLSCert() x509.Certificate {
	return x509.Certificate{
		Subject: pkix.Name{
			Organization: []string{"overchan"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Date(9005, 1, 1, 1, 1, 1, 1, time.UTC),
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA: true,
	}
}

// generate tls config, private key and certificate
func GenTLS(cfg *CryptoConfig) (tcfg *tls.Config, err error) {
	EnsureDir(cfg.cert_dir)
	// check for private key
	if !CheckFile(cfg.privkey_file) {
		// no private key, let's generate it
		log.Println("generating 4096 RSA private key...")
		k := newTLSCert()
		var priv *rsa.PrivateKey
		priv, err = rsa.GenerateKey(rand.Reader, 4096)
		if err == nil {
			serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 256)
			k.SerialNumber, err = rand.Int(rand.Reader, serialNumberLimit)
			k.DNSNames = append(k.DNSNames, cfg.hostname)
			k.Subject.CommonName = cfg.hostname
			if err == nil {
				var derBytes []byte
				derBytes, err = x509.CreateCertificate(rand.Reader, &k, &k, &priv.PublicKey, priv)
				var f io.WriteCloser
				f, err = os.Create(cfg.cert_file)
				if err == nil {
					err = pem.Encode(f, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
					f.Close()
					if err == nil {
						f, err = os.Create(cfg.privkey_file)
						if err == nil {
							err = pem.Encode(f, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
							f.Close()
						}
					}
				}
			}
		}
	}
	if err == nil {

		caPool := x509.NewCertPool()
		var m []string
		log.Println("checking", cfg.cert_dir, "for certificates")
		m, err = filepath.Glob(filepath.Join(cfg.cert_dir, "*.crt"))
		log.Println("loading", len(m), "trusted certificates")
		var data []byte
		for _, f := range m {
			var d []byte
			d, err = ioutil.ReadFile(f)
			if err == nil {
				data = append(data, d...)
			} else {
				return
			}
		}
		ok := caPool.AppendCertsFromPEM(data)
		if !ok {
			err = TlsFailedToLoadCA
			return
		}
		// we should have the key generated and stored by now
		var cert tls.Certificate
		cert, err = tls.LoadX509KeyPair(cfg.cert_file, cfg.privkey_file)
		if err == nil {
			tcfg = &tls.Config{
				CipherSuites: []uint16{tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384},
				RootCAs:      caPool,
				ClientCAs:    caPool,
				Certificates: []tls.Certificate{cert},
				ClientAuth:   tls.RequireAndVerifyClientCert,
			}
		}
	}
	return
}
