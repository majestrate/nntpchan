package cache

import (
	"io"
	"net/http"
)

// recache markup to io.Writer
type RecacheHandler func(io.Writer) error

type CacheInterface interface {
	ServeCached(w http.ResponseWriter, r *http.Request, key string, handler RecacheHandler)
	DeleteCache(key string)
	Cache(key string, body io.Reader)
	Has(key string) bool
	Close()
}
