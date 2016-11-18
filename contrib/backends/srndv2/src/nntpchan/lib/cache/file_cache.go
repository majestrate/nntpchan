// +build !disable_File

package cache

import (
	log "github.com/Sirupsen/logrus"
	"io"
	"net/http"
	"os"
	"time"
)

type FileCache struct {
}

func (self *FileCache) Has(key string) bool {
	_, err := os.Stat(key)
	return !os.IsNotExist(err)
}

func (self *FileCache) ServeCached(w http.ResponseWriter, r *http.Request, key string, handler RecacheHandler) {
	_, err := os.Stat(key)
	if os.IsNotExist(err) {
		modtime := time.Now().UTC()
		ts := modtime.Format(http.TimeFormat)

		w.Header().Set("Last-Modified", ts)
		f, err := os.Create(key)
		if err == nil {
			defer f.Close()
			mw := io.MultiWriter(f, w)
			err = handler(mw)
		}
		return
	}

	http.ServeFile(w, r, key)
}

func (self *FileCache) DeleteCache(key string) {
	err := os.Remove(key)
	if err != nil {
		log.Warnf("cannot remove file %s: %s", key, err.Error())
	}
}

func (self *FileCache) Cache(key string, body io.Reader) {
	f, err := os.Create(key)
	if err != nil {
		log.Warnf("cannot cache %s: %s", key, err.Error())
		return
	}
	defer f.Close()

	_, err = io.Copy(f, body)
	if err != nil {
		log.Warnf("cannot cache key %s: %s", key, err.Error())
	}
}

func (self *FileCache) Close() {
}

func NewFileCache() CacheInterface {
	cache := new(FileCache)

	return cache
}
