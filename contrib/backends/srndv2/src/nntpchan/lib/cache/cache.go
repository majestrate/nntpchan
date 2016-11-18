package cache

import (
	"github.com/majestrate/srndv2/lib/config"
	"strings"
)

// create cache from config structure
func FromConfig(c *config.CacheConfig) (cache CacheInterface, err error) {
	// set up cache
	if c != nil {
		// get cache backend
		cacheBackend := strings.ToLower(c.Backend)
		if cacheBackend == "redis" {
			// redis cache
			cache, err = NewRedisCache(c.Addr, c.Password)
		} else {
			// fall through
		}
	}
	if cache == nil {
		cache = NewNullCache()
	}
	return
}
