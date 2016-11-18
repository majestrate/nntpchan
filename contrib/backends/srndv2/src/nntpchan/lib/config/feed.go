package config

// configuration for 1 nntp feed
type FeedConfig struct {
	// feed's policy, filters articles
	Policy *ArticleConfig `json:"policy"`
	// remote server's address
	Addr string `json:"addr"`
	// proxy server config
	Proxy *ProxyConfig `json:"proxy"`
	// nntp username to log in with
	Username string `json:"username"`
	// nntp password to use when logging in
	Password string `json:"password"`
	// do we want to use tls?
	TLS bool `json:"tls"`
	// the name of this feed
	Name string `json:"name"`
	// how often to pull articles from the server in minutes
	// 0 for never
	PullInterval int `json:"pull"`
}

var DuummyFeed = FeedConfig{
	Policy: &DefaultArticlePolicy,
	Addr:   "nntp.dummy.tld:1119",
	Proxy:  &DefaultTorProxy,
	Name:   "dummy",
}

var DefaultFeeds = []*FeedConfig{
	&DuummyFeed,
}
