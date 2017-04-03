package srnd

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
)

type VarnishCache struct {
	varnish_url     string
	prefix          string
	handler         *nullHandler
	client          *http.Client
	regenThreadChan chan ArticleEntry
	regenGroupChan  chan groupRegenRequest
}

func (self *VarnishCache) invalidate(r string) {
	u, _ := url.Parse(r)
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
		go self.invalidate(fmt.Sprintf("%s%s%s-%d.html", self.varnish_url, self.prefix, group, n))
		go self.invalidate(fmt.Sprintf("%s%sb/%s/%d/", self.varnish_url, self.prefix, group, n))
		n--
	}
	self.invalidate(fmt.Sprintf("%s%sb/%s/", self.varnish_url, self.prefix, group))
}

// try to delete root post's page
func (self *VarnishCache) DeleteThreadMarkup(root_post_id string) {
	self.invalidate(fmt.Sprintf("%s%sthread-%s.html", self.varnish_url, self.prefix, HashMessageID(root_post_id)))
	self.invalidate(fmt.Sprintf("%s%st/%s/", self.varnish_url, self.prefix, HashMessageID(root_post_id)))
}

// regen every newsgroup
func (self *VarnishCache) RegenAll() {
	// we will do this as it's used by rengen on start for frontend
	groups := self.handler.database.GetAllNewsgroups()
	for _, group := range groups {
		self.handler.database.GetGroupThreads(group, self.regenThreadChan)
	}
}

func (self *VarnishCache) RegenFrontPage() {
	self.invalidate(fmt.Sprintf("%s%s", self.varnish_url, self.prefix))
}

func (self *VarnishCache) invalidateUkko() {
	// TODO: invalidate paginated ukko
	self.invalidate(fmt.Sprintf("%s%sukko.html", self.varnish_url, self.prefix))
	self.invalidate(fmt.Sprintf("%s%soverboard/", self.varnish_url, self.prefix))
	self.invalidate(fmt.Sprintf("%s%so/", self.varnish_url, self.prefix))
	// TODO: this is lazy af
	self.RegenFrontPage()
}

func (self *VarnishCache) pollRegen() {
	for {
		select {
		// consume regen requests
		case ev := <-self.regenGroupChan:
			{
				self.invalidate(fmt.Sprintf("%s%s%s-%d.html", self.varnish_url, self.prefix, ev.group, ev.page))
				self.invalidate(fmt.Sprintf("%s%sb/%s/%d/", self.varnish_url, self.prefix, ev.group, ev.page))
				if ev.page == 0 {
					self.invalidate(fmt.Sprintf("%s%sb/%s/", self.varnish_url, self.prefix, ev.group))
				}
			}
		case ev := <-self.regenThreadChan:
			{
				self.Regen(ev)
			}
		}
	}
}

// regen every page of the board
func (self *VarnishCache) RegenerateBoard(group string) {
	n, _ := self.handler.database.GetPagesPerBoard(group)
	for n > 0 {
		go self.invalidate(fmt.Sprintf("%s%s%s-%d.html", self.varnish_url, self.prefix, group, n))
		go self.invalidate(fmt.Sprintf("%s%s%s/%d/", self.varnish_url, self.prefix, group, n))
		n--
	}
	self.invalidate(fmt.Sprintf("%s%sb/%s/", self.varnish_url, self.prefix, group))
}

// regenerate pages after a mod event
func (self *VarnishCache) RegenOnModEvent(newsgroup, msgid, root string, page int) {
	self.regenGroupChan <- groupRegenRequest{newsgroup, page}
	self.regenThreadChan <- ArticleEntry{newsgroup, root}
}

func (self *VarnishCache) Start() {
	go self.pollRegen()
}

func (self *VarnishCache) Regen(msg ArticleEntry) {
	go self.invalidate(fmt.Sprintf("%s%s%s-%d.html", self.varnish_url, self.prefix, msg.Newsgroup(), 0))
	go self.invalidate(fmt.Sprintf("%s%s%s/%d/", self.varnish_url, self.prefix, msg.Newsgroup(), 0))
	go self.invalidate(fmt.Sprintf("%s%sthread-%s.html", self.varnish_url, self.prefix, HashMessageID(msg.MessageID())))
	go self.invalidate(fmt.Sprintf("%s%st/%s/", self.varnish_url, self.prefix, HashMessageID(msg.MessageID())))
	self.invalidateUkko()
}

func (self *VarnishCache) GetThreadChan() chan ArticleEntry {
	return self.regenThreadChan
}

func (self *VarnishCache) GetGroupChan() chan groupRegenRequest {
	return self.regenGroupChan
}

func (self *VarnishCache) GetHandler() http.Handler {
	return self.handler
}

func (self *VarnishCache) Close() {
	//nothig to do
}

func (self *VarnishCache) SetRequireCaptcha(required bool) {
	self.handler.requireCaptcha = required
}

func NewVarnishCache(varnish_url, bind_addr, prefix, webroot, name string, attachments bool, db Database, store ArticleStore) CacheInterface {
	cache := new(VarnishCache)
	cache.regenThreadChan = make(chan ArticleEntry, 16)
	cache.regenGroupChan = make(chan groupRegenRequest, 8)
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
	cache.prefix = prefix
	cache.handler = &nullHandler{
		prefix:         prefix,
		name:           name,
		attachments:    attachments,
		database:       db,
		requireCaptcha: true,
	}
	cache.varnish_url = varnish_url
	return cache
}
