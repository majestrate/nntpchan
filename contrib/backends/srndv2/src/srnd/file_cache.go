package srnd

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type FileCache struct {
	database Database
	store    ArticleStore

	webroot_dir string
	name        string

	regen_threads  int
	attachments    bool
	requireCaptcha bool

	prefix          string
	regenThreadChan chan ArticleEntry
	regenGroupChan  chan groupRegenRequest
	regenBoardMap   map[string]groupRegenRequest
	regenThreadMap  map[string]ArticleEntry
	regenCatalogMap map[string]bool

	regenBoardTicker   *time.Ticker
	ukkoTicker         *time.Ticker
	longTermTicker     *time.Ticker
	regenThreadTicker  *time.Ticker
	regenCatalogTicker *time.Ticker

	regenThreadLock  sync.RWMutex
	regenBoardLock   sync.RWMutex
	regenCatalogLock sync.RWMutex
}

func (self *FileCache) DeleteBoardMarkup(group string) {
	pages64 := self.database.GetGroupPageCount(group)
	pages := int(pages64)
	for page := 0; page < pages; page++ {
		fname := self.getFilenameForBoardPage(group, page, false)
		os.Remove(fname)
		fname = self.getFilenameForBoardPage(group, page, true)
		os.Remove(fname)
	}
}

// try to delete root post's page
func (self *FileCache) DeleteThreadMarkup(root_post_id string) {
	fname := self.getFilenameForThread(root_post_id, false)
	os.Remove(fname)
	fname = self.getFilenameForThread(root_post_id, true)
	os.Remove(fname)
}

func (self *FileCache) getFilenameForThread(root_post_id string, json bool) string {
	var ext string
	if json {
		ext = "json"
	} else {
		ext = "html"
	}
	fname := fmt.Sprintf("thread-%s.%s", HashMessageID(root_post_id), ext)
	return filepath.Join(self.webroot_dir, fname)
}

func (self *FileCache) getFilenameForBoardPage(boardname string, pageno int, json bool) string {
	var ext string
	if json {
		ext = "json"
	} else {
		ext = "html"
	}
	fname := fmt.Sprintf("%s-%d.%s", boardname, pageno, ext)
	return filepath.Join(self.webroot_dir, fname)
}

func (self *FileCache) getFilenameForCatalog(boardname string) string {
	fname := fmt.Sprintf("catalog-%s.html", boardname)
	return filepath.Join(self.webroot_dir, fname)
}

// regen every newsgroup
func (self *FileCache) RegenAll() {
	log.Println("regen all on http frontend")

	// get all groups
	groups := self.database.GetAllNewsgroups()
	if groups != nil {
		for _, group := range groups {
			// send every thread for this group down the regen thread channel
			go self.database.GetGroupThreads(group, self.regenThreadChan)
			pages := self.database.GetGroupPageCount(group)
			var pg int64
			for pg = 0; pg < pages; pg++ {
				self.regenGroupChan <- groupRegenRequest{group, int(pg)}
			}
		}
	}
}

func (self *FileCache) regenLongTerm() {
	wr, err := os.Create(filepath.Join(self.webroot_dir, "history.html"))
	defer wr.Close()
	if err != nil {
		log.Println("cannot render history graph", err)
		return
	}
	template.genGraphs(self.prefix, wr, self.database)
}

func (self *FileCache) pollLongTerm() {
	for {
		<-self.longTermTicker.C
		// regenerate long term stuff
		self.regenLongTerm()
	}
}

func (self *FileCache) pollRegen() {
	for {
		select {
		// listen for regen board requests
		case req := <-self.regenGroupChan:
			self.regenBoardLock.Lock()
			self.regenBoardMap[fmt.Sprintf("%s|%s", req.group, req.page)] = req
			self.regenBoardLock.Unlock()

			self.regenCatalogLock.Lock()
			self.regenCatalogMap[req.group] = true
			self.regenCatalogLock.Unlock()
			// listen for regen thread requests
		case entry := <-self.regenThreadChan:
			self.regenThreadLock.Lock()
			self.regenThreadMap[fmt.Sprintf("%s|%s", entry[0], entry[1])] = entry
			self.regenThreadLock.Unlock()
			// regen ukko
		case _ = <-self.ukkoTicker.C:
			self.regenUkko()
			self.RegenFrontPage()
		case _ = <-self.regenThreadTicker.C:
			self.regenThreadLock.Lock()
			for _, entry := range self.regenThreadMap {
				self.regenerateThread(entry, false)
				self.regenerateThread(entry, true)
			}
			self.regenThreadMap = make(map[string]ArticleEntry)
			self.regenThreadLock.Unlock()
		case _ = <-self.regenBoardTicker.C:
			self.regenBoardLock.Lock()
			for _, v := range self.regenBoardMap {
				self.regenerateBoardPage(v.group, v.page, false)
				self.regenerateBoardPage(v.group, v.page, true)
			}
			self.regenBoardMap = make(map[string]groupRegenRequest)
			self.regenBoardLock.Unlock()
		case _ = <-self.regenCatalogTicker.C:
			self.regenCatalogLock.Lock()
			for board, _ := range self.regenCatalogMap {
				self.regenerateCatalog(board)
			}
			self.regenCatalogMap = make(map[string]bool)
			self.regenCatalogLock.Unlock()
		}
	}
}

// regen every page of the board
func (self *FileCache) RegenerateBoard(group string) {
	pages, _ := self.database.GetPagesPerBoard(group)
	for page := 0; page < pages; page++ {
		self.regenerateBoardPage(group, page, false)
		self.regenerateBoardPage(group, page, true)
	}
}

// regenerate just a thread page
func (self *FileCache) regenerateThread(root ArticleEntry, json bool) {
	msgid := root.MessageID()
	if self.store.HasArticle(msgid) {
		fname := self.getFilenameForThread(msgid, json)
		wr, err := os.Create(fname)
		defer wr.Close()
		if err != nil {
			log.Println("did not write", fname, err)
			return
		}
		template.genThread(self.attachments, self.requireCaptcha, root, self.prefix, self.name, wr, self.database, json)
	} else {
		log.Println("don't have root post", msgid, "not regenerating thread")
	}
}

// regenerate just a page on a board
func (self *FileCache) regenerateBoardPage(board string, page int, json bool) {
	fname := self.getFilenameForBoardPage(board, page, json)
	wr, err := os.Create(fname)
	defer wr.Close()
	if err != nil {
		log.Println("error generating board page", page, "for", board, err)
		return
	}
	template.genBoardPage(self.attachments, self.requireCaptcha, self.prefix, self.name, board, page, wr, self.database, json)
}

// regenerate the catalog for a board
func (self *FileCache) regenerateCatalog(board string) {
	fname := self.getFilenameForCatalog(board)
	wr, err := os.Create(fname)
	defer wr.Close()
	if err != nil {
		log.Println("error generating catalog for", board, err)
		return
	}
	template.genCatalog(self.prefix, self.name, board, wr, self.database)
}

// regenerate the front page
func (self *FileCache) RegenFrontPage() {
	indexwr, err1 := os.Create(filepath.Join(self.webroot_dir, "index.html"))
	defer indexwr.Close()
	if err1 != nil {
		log.Println("cannot render front page", err1)
		return
	}
	boardswr, err2 := os.Create(filepath.Join(self.webroot_dir, "boards.html"))
	defer boardswr.Close()
	if err2 != nil {
		log.Println("cannot render board list page", err2)
		return
	}

	template.genFrontPage(10, self.prefix, self.name, indexwr, boardswr, self.database)

	j_boardswr, err2 := os.Create(filepath.Join(self.webroot_dir, "boards.json"))
	g := self.database.GetAllNewsgroups()
	err := json.NewEncoder(j_boardswr).Encode(g)
	if err != nil {
		log.Println("cannot render boards.json", err)
	}

}

// regenerate the overboard
func (self *FileCache) regenUkko() {

	// markup
	fname := filepath.Join(self.webroot_dir, "ukko.html")
	wr, err := os.Create(fname)
	defer wr.Close()
	if err != nil {
		log.Println("error generating ukko markup", err)
		return
	}
	template.genUkko(self.prefix, self.name, wr, self.database, false)

	// json
	fname = filepath.Join(self.webroot_dir, "ukko.json")
	wr, err = os.Create(fname)
	defer wr.Close()
	if err != nil {
		log.Println("error generating ukko json", err)
		return
	}
	template.genUkko(self.prefix, self.name, wr, self.database, true)
	i := 0
	for i < 10 {
		fname := fmt.Sprintf("ukko-%d.html", i)
		jname := fmt.Sprintf("ukko-%d.json", i)
		f, err := os.Create(fname)
		if err != nil {
			log.Println("Failed to create html ukko", i, err)
			return
		}
		defer f.Close()
		template.genUkkoPaginated(self.prefix, self.name, f, self.database, i, false)
		j, err := os.Create(jname)
		if err != nil {
			log.Printf("failed to create json ukko", i, err)
			return
		}
		defer j.Close()
		template.genUkkoPaginated(self.prefix, self.name, j, self.database, i, true)
	}
}

// regenerate pages after a mod event
func (self *FileCache) RegenOnModEvent(newsgroup, msgid, root string, page int) {
	if root == msgid {
		fname := self.getFilenameForThread(root, false)
		os.Remove(fname)
		fname = self.getFilenameForThread(root, true)
		os.Remove(fname)
	} else {
		self.regenThreadChan <- ArticleEntry{root, newsgroup}
	}
	self.regenGroupChan <- groupRegenRequest{newsgroup, int(page)}
	self.regenUkko()
}

func (self *FileCache) Start() {
	threads := self.regen_threads

	// check for invalid number of threads
	if threads <= 0 {
		threads = 1
	}

	// use N threads for regeneration
	for threads > 0 {
		go self.pollRegen()
		threads--
	}
	// run long term regen jobs
	go self.pollLongTerm()
}

func (self *FileCache) Regen(msg ArticleEntry) {
	self.regenThreadChan <- msg
	self.RegenerateBoard(msg.Newsgroup())
}

func (self *FileCache) GetThreadChan() chan ArticleEntry {
	return self.regenThreadChan
}

func (self *FileCache) GetGroupChan() chan groupRegenRequest {
	return self.regenGroupChan
}

func (self *FileCache) GetHandler() http.Handler {
	return http.FileServer(http.Dir(self.webroot_dir))
}

func (self *FileCache) Close() {
	//nothig to do
}

func (self *FileCache) SetRequireCaptcha(require bool) {
	self.requireCaptcha = require
}

func NewFileCache(prefix, webroot, name string, threads int, attachments bool, db Database, store ArticleStore) CacheInterface {
	cache := new(FileCache)

	cache.regenBoardTicker = time.NewTicker(time.Second * 10)
	cache.longTermTicker = time.NewTicker(time.Hour)
	cache.ukkoTicker = time.NewTicker(time.Second * 30)
	cache.regenThreadTicker = time.NewTicker(time.Second)
	cache.regenCatalogTicker = time.NewTicker(time.Second * 20)

	cache.regenBoardMap = make(map[string]groupRegenRequest)
	cache.regenThreadMap = make(map[string]ArticleEntry)
	cache.regenCatalogMap = make(map[string]bool)

	cache.regenThreadChan = make(chan ArticleEntry, 16)
	cache.regenGroupChan = make(chan groupRegenRequest, 8)

	cache.prefix = prefix
	cache.webroot_dir = webroot
	cache.name = name
	cache.regen_threads = threads
	cache.attachments = attachments
	cache.database = db
	cache.store = store

	return cache
}
