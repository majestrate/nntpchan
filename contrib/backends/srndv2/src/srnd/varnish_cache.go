package srnd

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
)

type VarnishCache struct {
	varnish_url      string
	prefix           string
	handler          *nullHandler
	client           *http.Client
	workers          int
	threadsRegenChan chan ArticleEntry
	invalidateChan   chan *url.URL
}

func (self *VarnishCache) InvertPagination() {
	self.handler.invertPagination = true
}

func (self *VarnishCache) invalidate(r string) {
	var langs []string
	langs = append(langs, "")
	self.handler.ForEachI18N(func(lang string) {
		langs = append(langs, lang)
	})
	for _, lang := range langs {
		u, _ := url.Parse(r)
		if lang != "" {
			q := u.Query()
			q.Add("lang", lang)
			u.RawQuery = q.Encode()
		}
		self.invalidateChan <- u
	}
}

func (self *VarnishCache) doRequest(u *url.URL) {
	if u == nil {
		return
	}
	resp, err := self.client.Do(&http.Request{
		Method: "PURGE",
		URL:    u,
	})
	if err == nil {
		resp.Body.Close()
	} else {
		log.Println("varnish cache error", err)
	}
}

func (self *VarnishCache) DeleteBoardMarkup(group string) {
	n, _ := self.handler.database.GetPagesPerBoard(group)
	for n > 0 {
		self.invalidate(fmt.Sprintf("%s%s%s-%d.html", self.varnish_url, self.prefix, group, n))
		self.invalidate(fmt.Sprintf("%s%sb/%s/%d/", self.varnish_url, self.prefix, group, n))
		n--
	}
	self.invalidate(fmt.Sprintf("%s%sb/%s/", self.varnish_url, self.prefix, group))
}

// try to delete root post's page
func (self *VarnishCache) DeleteThreadMarkup(root_post_id string) {
	id := HashMessageID(root_post_id)
	self.invalidate(fmt.Sprintf("%s%sthread-%s.json", self.varnish_url, self.prefix, id))
	self.invalidate(fmt.Sprintf("%s%st/%s/json", self.varnish_url, self.prefix, id))
	self.invalidate(fmt.Sprintf("%s%sthread-%s.html", self.varnish_url, self.prefix, id))
	self.invalidate(fmt.Sprintf("%s%st/%s/", self.varnish_url, self.prefix, id))
}

// regen every newsgroup
func (self *VarnishCache) RegenAll() {
	// we will do this as it's used by rengen on start for frontend
	groups := self.handler.database.GetAllNewsgroups()
	for _, group := range groups {
		self.handler.database.GetGroupThreads(group, self.threadsRegenChan)
	}
}

func (self *VarnishCache) RegenFrontPage() {
	self.invalidate(fmt.Sprintf("%s%s", self.varnish_url, self.prefix))
	// TODO: this is also lazy af
	self.invalidate(fmt.Sprintf("%s%shistory.html", self.varnish_url, self.prefix))
	self.invalidateUkko(10)
}

func (self *VarnishCache) invalidateUkko(pages int) {
	self.invalidate(fmt.Sprintf("%s%sukko.html", self.varnish_url, self.prefix))
	self.invalidate(fmt.Sprintf("%s%so/", self.varnish_url, self.prefix))
	self.invalidate(fmt.Sprintf("%s%sukko.json", self.varnish_url, self.prefix))
	self.invalidate(fmt.Sprintf("%s%so/json", self.varnish_url, self.prefix))
	n := 0
	for n < pages {
		self.invalidate(fmt.Sprintf("%s%so/%d/json", self.varnish_url, self.prefix, n))
		self.invalidate(fmt.Sprintf("%s%so/%d/", self.varnish_url, self.prefix, n))
		n++
	}
}

// regen every page of the board
func (self *VarnishCache) RegenerateBoard(group string) {
	n, _ := self.handler.database.GetPagesPerBoard(group)
	for n > 0 {
		self.invalidate(fmt.Sprintf("%s%s%s-%d.html", self.varnish_url, self.prefix, group, n))
		self.invalidate(fmt.Sprintf("%s%sb/%s/%d/", self.varnish_url, self.prefix, group, n))
		n--
	}
	self.invalidate(fmt.Sprintf("%s%sb/%s/", self.varnish_url, self.prefix, group))
}

// regenerate pages after a mod event
func (self *VarnishCache) RegenOnModEvent(newsgroup, msgid, root string, page int) {
	self.Regen(ArticleEntry{newsgroup, root})
	if page == 0 {
		self.invalidate(fmt.Sprintf("%s%sb/%s/", self.varnish_url, self.prefix, newsgroup))
	}
	self.invalidate(fmt.Sprintf("%s%sb/%s/%d/", self.varnish_url, self.prefix, newsgroup, page))
}

func (self *VarnishCache) poll() {
	for {
		ent := <-self.threadsRegenChan
		self.Regen(ent)
		self.RegenerateBoard(ent.Newsgroup())
	}
}

func (self *VarnishCache) Start() {
	go self.poll()
	workers := self.workers
	if workers <= 0 {
		workers = 1
	}
	for workers > 0 {
		go self.doWorker()
		workers--
	}
}

func (self *VarnishCache) doWorker() {
	for {
		self.doRequest(<-self.invalidateChan)
	}
}

func (self *VarnishCache) Regen(msg ArticleEntry) {
	self.DeleteThreadMarkup(msg.MessageID())
}

func (self *VarnishCache) GetHandler() CacheHandler {
	return self.handler
}

func (self *VarnishCache) Close() {
	//nothig to do
}

func (self *VarnishCache) SetRequireCaptcha(required bool) {
	self.handler.requireCaptcha = required
}

func NewVarnishCache(varnish_url, bind_addr, prefix, webroot, name, translations string, workers int, attachments bool, db Database, store ArticleStore) CacheInterface {
	cache := new(VarnishCache)
	cache.invalidateChan = make(chan *url.URL)
	cache.threadsRegenChan = make(chan ArticleEntry)
	cache.workers = workers
	local_addr, err := net.ResolveTCPAddr("tcp", bind_addr)
	if err != nil {
		log.Fatalf("failed to resolve %s for varnish cache: %s", bind_addr, err)
	}
	cache.client = &http.Client{
		Transport: &http.Transport{
			Dial: func(network, addr string) (c net.Conn, err error) {
				var remote_addr *net.TCPAddr
				remote_addr, err = net.ResolveTCPAddr(network, addr)
				if err == nil {
					c, err = net.DialTCP(network, local_addr, remote_addr)
				}
				return
			},
		},
	}
	cache.prefix = "/"
	cache.handler = &nullHandler{
		prefix:         prefix,
		name:           name,
		attachments:    attachments,
		database:       db,
		requireCaptcha: true,
		i18n:           make(map[string]*I18N),
		translations:   translations,
	}
	cache.varnish_url = varnish_url
	return cache
}
