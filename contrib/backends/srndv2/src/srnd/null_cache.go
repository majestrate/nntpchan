package srnd

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
)

type NullCache struct {
	regenThreadChan chan ArticleEntry
	regenGroupChan  chan groupRegenRequest
	handler         *nullHandler
}

type nullHandler struct {
	database       Database
	attachments    bool
	requireCaptcha bool
	name           string
	prefix         string
}

func (self *nullHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	_, file := filepath.Split(path)

	isjson := strings.HasSuffix(path, "/json") || strings.HasSuffix(path, "/json/")

	if strings.HasPrefix(path, "/t/") {
		// thread handler
		parts := strings.Split(path[3:], "/")
		hash := parts[0]
		msg, err := self.database.GetMessageIDByHash(hash)
		if err == nil {
			template.genThread(self.attachments, self.requireCaptcha, msg, self.prefix, self.name, w, self.database, isjson)
			return
		} else {
			goto notfound
		}
	}
	if strings.Trim(path, "/") == "overboard" {
		// generate ukko aka overboard
		template.genUkko(self.prefix, self.name, w, self.database, isjson)
		return
	}

	if strings.HasPrefix(path, "/b/") {
		// board handler
		parts := strings.Split(path[3:], "/")
		page := 0
		group := parts[0]
		if len(parts) > 1 && parts[1] != "" && parts[1] != "json" {
			var err error
			page, err = strconv.Atoi(parts[1])
			if err != nil {
				goto notfound
			}
		}
		hasgroup := self.database.HasNewsgroup(group)
		if !hasgroup {
			goto notfound
		}
		pages := self.database.GetGroupPageCount(group)
		if page >= int(pages) {
			goto notfound
		}
		template.genBoardPage(self.attachments, self.requireCaptcha, self.prefix, self.name, group, page, w, self.database, isjson)
		return
	}

	if strings.HasPrefix(path, "/o/") {
		page := 0
		parts := strings.Split(path[3:], "/")
		if parts[0] != "json" && parts[0] != "" {
			var err error
			page, err = strconv.Atoi(parts[0])
			if err != nil {
				goto notfound
			}
		}
		template.genUkkoPaginated(self.prefix, self.name, w, self.database, page, isjson)
		return
	}

	if len(file) == 0 || file == "index.html" {
		template.genFrontPage(10, self.prefix, self.name, w, ioutil.Discard, self.database)
		return
	}

	if file == "index.json" {
		// TODO: index.json
		goto notfound
	}
	if strings.HasPrefix(file, "history.html") {
		template.genGraphs(self.prefix, w, self.database)
		return
	}
	if strings.HasPrefix(file, "boards.html") {
		template.genFrontPage(10, self.prefix, self.name, ioutil.Discard, w, self.database)
		return
	}

	if strings.HasPrefix(file, "boards.json") {
		b := self.database.GetAllNewsgroups()
		json.NewEncoder(w).Encode(b)
		return
	}

	if strings.HasPrefix(file, "ukko.html") {
		template.genUkko(self.prefix, self.name, w, self.database, false)
		return
	}
	if strings.HasPrefix(file, "ukko.json") {
		template.genUkko(self.prefix, self.name, w, self.database, true)
		return
	}

	if strings.HasPrefix(file, "ukko-") {
		page := getUkkoPage(file)
		template.genUkkoPaginated(self.prefix, self.name, w, self.database, page, isjson)
		return
	}
	if strings.HasPrefix(file, "thread-") {
		hash := getThreadHash(file)
		if len(hash) == 0 {
			goto notfound
		}
		msg, err := self.database.GetMessageIDByHash(hash)
		if err != nil {
			goto notfound
		}
		template.genThread(self.attachments, self.requireCaptcha, msg, self.prefix, self.name, w, self.database, isjson)
		return
	}
	if strings.HasPrefix(file, "catalog-") {
		group := getGroupForCatalog(file)
		if len(group) == 0 {
			goto notfound
		}
		hasgroup := self.database.HasNewsgroup(group)
		if !hasgroup {
			goto notfound
		}
		template.genCatalog(self.prefix, self.name, group, w, self.database)
		return
	} else {
		group, page := getGroupAndPage(file)
		if len(group) == 0 || page < 0 {
			goto notfound
		}
		hasgroup := self.database.HasNewsgroup(group)
		if !hasgroup {
			goto notfound
		}
		pages := self.database.GetGroupPageCount(group)
		if page >= int(pages) {
			goto notfound
		}
		template.genBoardPage(self.attachments, self.requireCaptcha, self.prefix, self.name, group, page, w, self.database, isjson)
		return
	}

notfound:
	template.renderNotFound(w, r, self.prefix, self.name)
}

func (self *NullCache) DeleteBoardMarkup(group string) {
}

// try to delete root post's page
func (self *NullCache) DeleteThreadMarkup(root_post_id string) {
}

// regen every newsgroup
func (self *NullCache) RegenAll() {
}

func (self *NullCache) RegenFrontPage() {
}

func (self *NullCache) SetRequireCaptcha(required bool) {
	self.handler.requireCaptcha = required
}

func (self *NullCache) pollRegen() {
	for {
		select {
		// consume regen requests
		case _ = <-self.regenGroupChan:
			{
			}
		case _ = <-self.regenThreadChan:
			{
			}
		}
	}
}

// regen every page of the board
func (self *NullCache) RegenerateBoard(group string) {
}

// regenerate pages after a mod event
func (self *NullCache) RegenOnModEvent(newsgroup, msgid, root string, page int) {
}

func (self *NullCache) Start() {
	go self.pollRegen()
}

func (self *NullCache) Regen(msg ArticleEntry) {
}

func (self *NullCache) GetThreadChan() chan ArticleEntry {
	return self.regenThreadChan
}

func (self *NullCache) GetGroupChan() chan groupRegenRequest {
	return self.regenGroupChan
}

func (self *NullCache) GetHandler() http.Handler {
	return self.handler
}

func (self *NullCache) Close() {
	//nothig to do
}

func NewNullCache(prefix, webroot, name string, attachments bool, db Database, store ArticleStore) CacheInterface {
	cache := new(NullCache)
	cache.regenThreadChan = make(chan ArticleEntry, 16)
	cache.regenGroupChan = make(chan groupRegenRequest, 8)
	cache.handler = &nullHandler{
		prefix:         prefix,
		name:           name,
		attachments:    attachments,
		requireCaptcha: true,
		database:       db,
	}

	return cache
}
