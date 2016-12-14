package config

// settings for setting up ssl
type SSLSettings struct {
	// path to ssl private key
	SSLKeyFile string `json:"key"`
	// path to ssl certificate signed by CA
	SSLCertFile string `json:"cert"`
	// domain name to use for ssl
	DomainName string `json:"fqdn"`
}
