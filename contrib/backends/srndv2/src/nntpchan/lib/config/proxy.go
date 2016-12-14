package config

// proxy configuration
type ProxyConfig struct {
	Type string `json:"type"`
	Addr string `json:"addr"`
}

// default tor proxy
var DefaultTorProxy = ProxyConfig{
	Type: "socks",
	Addr: "127.0.0.1:9050",
}
