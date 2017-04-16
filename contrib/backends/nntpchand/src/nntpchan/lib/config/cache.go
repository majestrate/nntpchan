package config

// caching interface configuration
type CacheConfig struct {
	// backend cache driver name
	Backend string `json:"backend"`
	// address for cache
	Addr string `json:"addr"`
	// username for login
	User string `json:"user"`
	// password for login
	Password string `json:"password"`
}
