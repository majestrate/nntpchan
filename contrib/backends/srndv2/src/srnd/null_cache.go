package srnd

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

type NullCache struct {
	handler *nullHandler
}

func (self *NullCache) InvertPagination() {
	self.handler.invertPagination = true
}

type nullHandler struct {
	database         Database
	attachments      bool
	requireCaptcha   bool
	name             string
	prefix           string
	translations     string
	i18n             map[string]*I18N
	access           sync.Mutex
	invertPagination bool
}

func (self *nullHandler) ForEachI18N(v func(string)) {
	self.access.Lock()
	for lang := range self.i18n {
		v(lang)
	}
	self.access.Unlock()
}

func (self *nullHandler) GetI18N(r *http.Request) *I18N {
	lang := r.URL.Query().Get("lang")
	if lang == "" {
		lang = I18nProvider.Name
	}
	self.access.Lock()
	i, ok := self.i18n[lang]
	if !ok {
		var err error
		i, err = NewI18n(lang, self.translations)
		if err != nil {
			log.Println(err)
		}
		if i != nil {
			self.i18n[lang] = i
		}
	}
	self.access.Unlock()
	return i
}

func (self *nullHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	i18n := self.GetI18N(r)

	path := r.URL.Path
	_, file := filepath.Split(path)

	isjson := strings.HasSuffix(path, "/json") || strings.HasSuffix(path, "/json/")

	if strings.HasPrefix(path, "/t/") {
		// thread handler
		parts := strings.Split(path[3:], "/")
		hash := parts[0]
		msg, err := self.database.GetMessageIDByHash(hash)
		if err == nil {

			if !self.database.HasArticleLocal(msg.MessageID()) {
				goto notfound
			}

			template.genThread(self.attachments, self.requireCaptcha, msg, self.prefix, self.name, w, self.database, isjson, i18n)
			return
		} else {
			goto notfound
		}
	}
	if strings.Trim(path, "/") == "overboard" {
		// generate ukko aka overboard
		template.genUkko(self.prefix, self.name, w, self.database, isjson, i18n, self.invertPagination)
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
		if group == "" {
			goto notfound
		}
		hasgroup := self.database.HasNewsgroup(group)
		if !hasgroup {
			goto notfound
		}

		banned, _ := self.database.NewsgroupBanned(group)
		if banned {
			goto notfound
		}

		pages := self.database.GetGroupPageCount(group)
		if page >= int(pages) {
			goto notfound
		}

		template.genBoardPage(self.attachments, self.requireCaptcha, self.prefix, self.name, group, int(pages), page, w, self.database, isjson, i18n, self.invertPagination)
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
		pages, _ := self.database.GetUkkoPageCount(10)
		if path == "/o/" {
			if self.invertPagination {
				page = int(pages)
			}
		}
		template.genUkkoPaginated(self.prefix, self.name, w, self.database, int(pages), page, isjson, i18n, self.invertPagination)
		return
	}

	if len(file) == 0 || file == "index.html" {
		template.genFrontPage(10, self.prefix, self.name, w, ioutil.Discard, self.database, i18n)
		return
	}

	if file == "index.json" {
		// TODO: index.json
		goto notfound
	}
	if strings.HasPrefix(file, "history.html") {
		template.genGraphs(self.prefix, w, self.database, i18n)
		return
	}
	if strings.HasPrefix(file, "boards.html") {
		template.genBoardList(self.prefix, self.name, w, self.database, i18n)
		return
	}

	if strings.HasPrefix(file, "boards.json") {
		b := self.database.GetAllNewsgroups()
		json.NewEncoder(w).Encode(b)
		return
	}

	if strings.HasPrefix(file, "ukko.html") {
		template.genUkko(self.prefix, self.name, w, self.database, false, i18n, self.invertPagination)
		return
	}
	if strings.HasPrefix(file, "ukko.json") {
		template.genUkko(self.prefix, self.name, w, self.database, true, i18n, self.invertPagination)
		return
	}

	if strings.HasPrefix(file, "ukko-") {
		page := getUkkoPage(file)
		pages, _ := self.database.GetUkkoPageCount(10)
		template.genUkkoPaginated(self.prefix, self.name, w, self.database, int(pages), page, isjson, i18n, self.invertPagination)
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

		if !self.database.HasArticleLocal(msg.MessageID()) {
			goto notfound
		}

		template.genThread(self.attachments, self.requireCaptcha, msg, self.prefix, self.name, w, self.database, isjson, i18n)
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
		template.genCatalog(self.prefix, self.name, group, w, self.database, i18n)
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
		template.genBoardPage(self.attachments, self.requireCaptcha, self.prefix, self.name, group, int(pages), page, w, self.database, isjson, i18n, self.invertPagination)
		return
	}

notfound:
	template.renderNotFound(w, r, self.prefix, self.name, i18n)
}

func (self *NullCache) DeleteBoardMarkup(group string) {
}

// try to delete root post's page
func (self *NullCache) DeleteThreadMarkup(root_post_id string) {
}

// regen every newsgroup
func (self *NullCache) RegenAll() {
}

func (self *NullCache) RegenFrontPage(pagestart int) {
}

func (self *NullCache) SetRequireCaptcha(required bool) {
	self.handler.requireCaptcha = required
}

// regen every page of the board
func (self *NullCache) RegenerateBoard(group string) {
}

// regenerate pages after a mod event
func (self *NullCache) RegenOnModEvent(newsgroup, msgid, root string, page int) {
}

func (self *NullCache) Start() {
}

func (self *NullCache) Regen(msg ArticleEntry) {
}

func (self *NullCache) GetHandler() CacheHandler {
	return self.handler
}

func (self *NullCache) Close() {
	//nothig to do
}

func NewNullCache(prefix, webroot, name, translations string, attachments bool, db Database, store ArticleStore) CacheInterface {
	cache := new(NullCache)
	cache.handler = &nullHandler{
		prefix:         prefix,
		name:           name,
		attachments:    attachments,
		requireCaptcha: true,
		database:       db,
		i18n:           make(map[string]*I18N),
		translations:   translations,
	}

	return cache
}
