package cache

import (
	"io"
	"io/ioutil"
	"net/http"
)

type NullCache struct {
}

func (self *NullCache) ServeCached(w http.ResponseWriter, r *http.Request, key string, handler RecacheHandler) {
	handler(w)
}

func (self *NullCache) DeleteCache(key string) {
}

func (self *NullCache) Cache(key string, body io.Reader) {
	io.Copy(ioutil.Discard, body)
}

func (self *NullCache) Close() {
}

func (self *NullCache) Has(key string) bool {
	return false
}

func NewNullCache() CacheInterface {
	cache := new(NullCache)
	return cache
}
