//
// daemon.go
//
package srnd

import (
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"net/textproto"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// the state of a feed that we are persisting
type feedState struct {
	Config  *FeedConfig
	Paused  bool
	Exiting bool
}

// the status of a feed that we are persisting
type feedStatus struct {
	// does this feed exist?
	Exists bool
	// the active connections this feed has open if it exists
	Conns []*nntpConnection
	// the state of this feed if it exists
	State *feedState
}

// an event for querying if a feed's status
type feedStatusQuery struct {
	// name of feed
	name string
	// channel to send result down
	resultChnl chan *feedStatus
}

// the result of modifying a feed
type modifyFeedPolicyResult struct {
	// error if one occured
	// set to nil if no error occured
	err error
	// name of the feed that was changed
	// XXX: is this needed?
	name string
}

// describes how we want to change a feed's policy
type modifyFeedPolicyEvent struct {
	// name of feed
	name string
	// new policy
	policy FeedPolicy
	// channel to send result down
	// if nil don't send result
	resultChnl chan *modifyFeedPolicyResult
}

type NNTPDaemon struct {
	instance_name string
	bind_addr     string
	conf          *SRNdConfig
	store         ArticleStore
	database      Database
	mod           ModEngine
	expire        ExpirationCore
	listener      net.Listener
	debug         bool
	sync_on_start bool
	// anon settings
	allow_anon             bool
	allow_anon_attachments bool

	// do we allow attachments from remote?
	allow_attachments bool

	running bool
	// http frontend
	frontend Frontend

	//cache driver
	cache CacheInterface

	// current feeds loaded from config
	loadedFeeds map[string]*feedState
	// for obtaining a list of loaded feeds from the daemon
	get_feeds chan chan []*feedStatus
	// for obtaining the status of a loaded feed
	get_feed chan *feedStatusQuery
	// for modifying feed's policies
	modify_feed_policy chan *modifyFeedPolicyEvent
	// for registering a new feed to persist
	register_feed chan FeedConfig
	// for degregistering an existing feed from persistance given name
	deregister_feed chan string
	// map of name -> NNTPConnection
	activeConnections map[string]*nntpConnection
	// for registering and deregistering outbound feed connections
	register_connection   chan *nntpConnection
	deregister_connection chan *nntpConnection

	// channel for broadcasting a message to all feeds given their newsgroup, message_id
	send_all_feeds chan ArticleEntry
	// channel for broadcasting an ARTICLE command to all feeds in reader mode
	ask_for_article chan string
	// operation of daemon done after sending bool down this channel
	done chan bool

	tls_config *tls.Config

	send_articles_mtx sync.RWMutex
	send_articles     []ArticleEntry
	ask_articles_mtx  sync.RWMutex
	ask_articles      []string

	pump_ticker       *time.Ticker
	expiration_ticker *time.Ticker
	article_lifetime  time.Duration

	spamFilter SpamFilter
}

// return true if text passes all checks and is okay for posting
func (self *NNTPDaemon) CheckText(text string) bool {
	for _, re := range self.conf.filter.globalFilters {
		if re.MatchString(text) {
			return false
		}
	}
	return true
}

func (self NNTPDaemon) End() {
	if self.listener != nil {
		self.listener.Close()
	}
	if self.database != nil {
		self.database.Close()
	}
	if self.cache != nil {
		self.cache.Close()
	}
	self.done <- true
}

func (self *NNTPDaemon) GetDatabase() Database {
	return self.database
}

// sign an article as coming from our daemon
func (self *NNTPDaemon) WrapSign(nntp NNTPMessage) {
	sk, ok := self.conf.daemon["secretkey"]
	if ok {
		seed := parseTripcodeSecret(sk)
		if seed == nil {
			log.Println("invalid secretkey will not sign")
		} else {
			pk, sec := naclSeedToKeyPair(seed)
			sig := msgidFrontendSign(sec, nntp.MessageID())
			nntp.Headers().Add("X-Frontend-Signature", sig)
			nntp.Headers().Add("X-Frontend-Pubkey", hexify(pk))
		}
	} else {
		log.Println("sending", nntp.MessageID(), "unsigned")
	}
}

// for srnd tool
func (self *NNTPDaemon) DelNNTPLogin(username string) {
	exists, err := self.database.CheckNNTPUserExists(username)
	if !exists {
		log.Println("user", username, "does not exist")
		return
	} else if err == nil {
		err = self.database.RemoveNNTPLogin(username)
	}
	if err == nil {
		log.Println("removed user", username)
	} else {
		log.Fatalf("error removing nntp login: %s", err.Error())
	}
}

// for srnd tool
func (self *NNTPDaemon) AddNNTPLogin(username, password string) {
	exists, err := self.database.CheckNNTPUserExists(username)
	if exists {
		log.Println("user", username, "exists")
		return
	} else if err == nil {
		err = self.database.AddNNTPLogin(username, password)
	}
	if err == nil {
		log.Println("added user", username)
	} else {
		log.Fatalf("error adding nntp login: %s", err.Error())
	}
}

func (self *NNTPDaemon) dialOut(proxy_type, proxy_addr, remote_addr string) (conn net.Conn, err error) {

	if proxy_type == "" || proxy_type == "none" {
		// connect out without proxy
		log.Println("dial out to ", remote_addr)
		conn, err = net.Dial("tcp", remote_addr)
		if err != nil {
			log.Println("cannot connect to outfeed", remote_addr, err)
			return
		}
	} else if proxy_type == "socks4a" || proxy_type == "socks" {
		// connect via socks4a
		log.Println("dial out via proxy", proxy_addr)
		conn, err = net.Dial("tcp", proxy_addr)
		if err != nil {
			log.Println("cannot connect to proxy", proxy_addr)
			return
		}
		// generate request
		idx := strings.LastIndex(remote_addr, ":")
		if idx == -1 {
			err = errors.New("invalid address: " + remote_addr)
			return
		}
		var port uint64
		addr := remote_addr[:idx]
		port, err = strconv.ParseUint(remote_addr[idx+1:], 10, 16)
		if port >= 25536 {
			err = errors.New("bad proxy port")
			return
		} else if err != nil {
			return
		}
		var proxy_port uint16
		proxy_port = uint16(port)
		proxy_ident := "srndv2"
		req_len := len(addr) + 1 + len(proxy_ident) + 1 + 8

		req := make([]byte, req_len)
		// pack request
		req[0] = '\x04'
		req[1] = '\x01'
		req[2] = byte(proxy_port & 0xff00 >> 8)
		req[3] = byte(proxy_port & 0x00ff)
		req[7] = '\x01'
		idx = 8

		proxy_ident_b := []byte(proxy_ident)
		addr_b := []byte(addr)

		var bi int
		for bi = range proxy_ident_b {
			req[idx] = proxy_ident_b[bi]
			idx += 1
		}
		idx += 1
		for bi = range addr_b {
			req[idx] = addr_b[bi]
			idx += 1
		}

		log.Println("dial out via proxy", proxy_addr)
		conn, err = net.Dial("tcp", proxy_addr)
		// send request
		_, err = conn.Write(req)
		resp := make([]byte, 8)

		// receive response
		_, err = conn.Read(resp)
		if resp[1] == '\x5a' {
			// success
			log.Println("connected to", addr)
		} else {
			log.Println("failed to connect to", addr)
			conn.Close()
			conn = nil
			err = errors.New("failed to connect via proxy")
			return
		}
	} else if proxy_type == "http" {
		log.Println("dial out via http proxy", proxy_addr)
		conn, err = net.Dial("tcp", proxy_addr)
		if err == nil {
			_, err = fmt.Fprintf(conn, "CONNECT %s HTTP/1.1\r\n\r\n", remote_addr)
			if err == nil {
				readLine := func(c net.Conn) (line string, e error) {
					var buff [1]byte
					var n int
					for e == nil {
						n, e = c.Read(buff[:])
						if n > 0 {
							line += string(buff[:])
							if buff[0] == 10 {
								return
							}
						}
					}
					return
				}
				var line string
				line, err = readLine(conn)
				if strings.HasPrefix(line, "HTTP/1.1 200") || strings.HasPrefix(line, "HTTP/1.0 200") {
					log.Println("http proxy connect accepted")
					for err == nil {
						line, err = readLine(conn)
						if line == "\r\n" {
							break
						}
					}
				} else {
					err = errors.New("proxy request rejected: " + strings.Trim(line, "\r"))
					log.Println(err)
					conn.Close()
					conn = nil
				}
			}
		}
	} else {
		err = errors.New("invalid proxy type: " + proxy_type)
	}
	return
}

// save current feeds to feeds.ini, overwrites feeds.ini
// returns error if one occurs while writing to feeds.ini
func (self *NNTPDaemon) storeFeedsConfig() (err error) {
	feeds := self.activeFeeds()
	var feedconfigs []FeedConfig
	for _, status := range feeds {
		feedconfigs = append(feedconfigs, *status.State.Config)
	}
	err = SaveFeeds(feedconfigs, self.conf.inboundPolicy)
	return
}

func (self *NNTPDaemon) AllowsNewsgroup(group string) bool {
	return self.conf.inboundPolicy == nil || self.conf.inboundPolicy.AllowsNewsgroup(group)
}

// change a feed's policy given the feed's name
// return error if one occurs while modifying feed's policy
func (self *NNTPDaemon) modifyFeedPolicy(feedname string, policy FeedPolicy) (err error) {
	// make event
	chnl := make(chan *modifyFeedPolicyResult)
	ev := &modifyFeedPolicyEvent{
		resultChnl: chnl,
		name:       feedname,
		policy:     policy,
	}
	// fire event
	self.modify_feed_policy <- ev
	// recv result
	result := <-chnl
	if result == nil {
		// XXX: why would this ever happen?
		err = errors.New("no result from daemon after modifying feed")
	} else {
		err = result.err
	}
	// done with the event result channel
	close(chnl)
	return
}

// remove a persisted feed from the daemon
// does not modify feeds.ini
func (self *NNTPDaemon) removeFeed(feedname string) (err error) {
	// deregister feed first so it doesn't reconnect immediately
	self.deregister_feed <- feedname
	return
}

func (self *NNTPDaemon) getFeedStatus(feedname string) (status *feedStatus) {
	chnl := make(chan *feedStatus)
	self.get_feed <- &feedStatusQuery{
		name:       feedname,
		resultChnl: chnl,
	}
	status = <-chnl
	close(chnl)
	return
}

// add a feed to be persisted by the daemon
// does not modify feeds.ini
func (self *NNTPDaemon) addFeed(conf FeedConfig) (err error) {
	self.register_feed <- conf
	return
}

// get an immutable list of all active feeds
func (self *NNTPDaemon) activeFeeds() (feeds []*feedStatus) {
	chnl := make(chan []*feedStatus)
	// query feeds
	self.get_feeds <- chnl
	// get reply
	feeds = <-chnl
	// got reply, close channel
	close(chnl)
	return
}

func (self *NNTPDaemon) messageSizeLimitFor(newsgroup string) int64 {
	// TODO: per newsgroup
	return mapGetInt64(self.conf.store, "max_message_size", DefaultMaxMessageSize)
}

func (self *NNTPDaemon) persistFeed(conf *FeedConfig, mode string, n int) {
	if conf.disable {
		log.Println(conf.Name, "is disabled not persisting")
		return
	}
	log.Println(conf.Name, "persisting in", mode, "mode")
	backoff := time.Second
	for {
		if self.running {
			// get the status of this feed
			status := self.getFeedStatus(conf.Name)
			if !status.Exists {
				// our feed was removed
				// let's die
				log.Println(conf.Name, "ended", mode, "mode")
				return
			}

			if status.State.Paused {
				// we are paused
				// sleep for a bit
				time.Sleep(time.Second)
				// check status again
				continue
			}
			// do we want to do a pull based sync?

			if mode == "sync" {
				// yeh, do it
				self.syncPull(conf)
				// sleep for the sleep interval and continue
				log.Println(conf.Name, "waiting for", conf.sync_interval, "before next sync")
				time.Sleep(conf.sync_interval)
				continue
			}
			conn, err := self.dialOut(conf.proxy_type, conf.proxy_addr, conf.Addr)
			if err != nil {
				log.Println(conf.Name, "failed to dial out", err.Error())
				log.Println(conf.Name, "back off for", backoff, "seconds")
				time.Sleep(backoff)
				// exponential backoff
				if backoff < (10 * time.Minute) {
					backoff *= 2
				}
				continue
			}
			nntp := createNNTPConnection(conf.Addr)
			nntp.policy = &conf.policy
			nntp.feedname = conf.Name
			nntp.name = fmt.Sprintf("%s-%d-%s", conf.Name, n, mode)
			stream, reader, use_tls, err := nntp.outboundHandshake(textproto.NewConn(conn), conf)
			if err == nil {
				if mode == "reader" && !reader {
					log.Println(nntp.name, "we don't support reader on this feed, dropping")
					conn.Close()
				} else {
					self.register_connection <- nntp
					// success connecting, reset backoff
					backoff = time.Second
					// run connection
					nntp.runConnection(self, false, stream, reader, use_tls, mode, conn, conf)
					// deregister connection
					self.deregister_connection <- nntp
				}
			} else {
				log.Println("error doing outbound hanshake", err)
			}
		}
		log.Println(conf.Name, "back off for", backoff, "seconds")
		time.Sleep(backoff)
		// exponential backoff
		if backoff < (10 * time.Minute) {
			backoff *= 2
		}
	}
}

// do a oneshot pull based sync with another server
func (self *NNTPDaemon) syncPull(conf *FeedConfig) {
	c, err := self.dialOut(conf.proxy_type, conf.proxy_addr, conf.Addr)
	if err == nil {
		conn := textproto.NewConn(c)
		// we connected
		nntp := createNNTPConnection(conf.Addr)
		nntp.name = conf.Addr + "-sync"
		nntp.feedname = conf.Name
		// do handshake
		_, reader, _, err := nntp.outboundHandshake(conn, conf)

		if err != nil {
			log.Println("failed to scrape server", err)
		}
		if reader {
			// we can do it
			err = nntp.scrapeServer(self, conn)
			if err == nil { // we succeeded
				log.Println(nntp.name, "Scrape successful")
				nntp.Quit(conn)
				conn.Close()
			} else {
				// we failed
				log.Println(nntp.name, "scrape failed", err)
				conn.Close()
			}
		} else if err == nil {
			// we can't do it
			log.Println(nntp.name, "does not support reader mode, cancel scrape")
			nntp.Quit(conn)
		} else {
			// error happened
			log.Println(nntp.name, "error occurred when scraping", err)
		}
	}
}

func (self *NNTPDaemon) ExpireAll() {
	log.Println("expiring all orphans")
	self.expire = createExpirationCore(self.database, self.store, self.informHooks)
	self.expire.ExpireOrphans()
}

func (self *NNTPDaemon) MarkSpam(msgid string) {
	if ValidMessageID(msgid) {
		err := self.mod.MarkSpam(msgid)
		if err != nil {
			log.Println(err)
		}
	}
}

// run daemon
func (self *NNTPDaemon) Run() {
	self.bind_addr = self.conf.daemon["bind"]

	listener, err := net.Listen("tcp", self.bind_addr)
	if err != nil {
		log.Fatal("failed to bind to", self.bind_addr, err)
	}
	self.listener = listener
	log.Printf("SRNd NNTPD bound at %s", listener.Addr())

	if self.conf.pprof != nil && self.conf.pprof.enable {
		addr := self.conf.pprof.bind
		log.Println("pprof enabled, binding to", addr)
		go func() {
			err := http.ListenAndServe(addr, nil)
			if err != nil {
				log.Fatalf("error from pprof, RIP srndv2: %s", err.Error())
			}
		}()
	}

	// write pid file
	pidfile := self.conf.daemon["pidfile"]
	if pidfile != "" {
		f, err := os.OpenFile(pidfile, os.O_CREATE|os.O_WRONLY, 0644)
		if err == nil {
			pid := os.Getpid()
			fmt.Fprintf(f, "%d", pid)
			f.Close()
		} else {
			log.Fatalf("failed to open pidfile %s: %s", pidfile, err)
		}
	}

	self.register_connection = make(chan *nntpConnection)
	self.deregister_connection = make(chan *nntpConnection)
	self.send_all_feeds = make(chan ArticleEntry, 128)
	self.activeConnections = make(map[string]*nntpConnection)
	self.loadedFeeds = make(map[string]*feedState)
	self.register_feed = make(chan FeedConfig)
	self.deregister_feed = make(chan string)
	self.get_feeds = make(chan chan []*feedStatus)
	self.get_feed = make(chan *feedStatusQuery)
	self.modify_feed_policy = make(chan *modifyFeedPolicyEvent)
	self.ask_for_article = make(chan string, 128)

	self.pump_ticker = time.NewTicker(time.Millisecond * 100)
	if self.conf.daemon["archive"] == "1" {
		log.Println("running in archive mode")
		self.expire = nil
		self.frontend.ArchiveMode()
	} else {
		self.expire = createExpirationCore(self.database, self.store, self.informHooks)
	}
	self.sync_on_start = self.conf.daemon["sync_on_start"] == "1"
	self.instance_name = self.conf.daemon["instance_name"]
	self.allow_anon = self.conf.daemon["allow_anon"] == "1"
	self.allow_anon_attachments = self.conf.daemon["allow_anon_attachments"] == "1"
	self.allow_attachments = self.conf.daemon["allow_attachments"] == "1"

	// set up admin user if it's specified in the config
	pubkey, ok := self.conf.frontend["admin_key"]
	if ok {
		// TODO: check for valid format
		var isadmin bool
		isadmin, err := self.database.CheckAdminPubkey(pubkey)
		if !isadmin {
			log.Println("add admin key", pubkey)
			err = self.database.MarkPubkeyAdmin(pubkey)
		}
		if err != nil {
			log.Printf("failed to add admin mod key, %s", err)
		}
	}

	// start frontend
	go self.frontend.Mainloop()

	log.Println("we have", len(self.conf.feeds), "feeds")

	defer self.listener.Close()
	// run expiration mainloop
	if self.expire == nil {
		log.Println("we are an archive, not expiring posts")
	} else {
		lifetime := mapGetInt(self.conf.daemon, "article_lifetime", 0)
		if lifetime > 0 {
			self.article_lifetime = time.Duration(lifetime) * time.Hour
			since := 0 - (self.article_lifetime)
			self.expire.ExpireBefore(time.Now().Add(since))
			self.expiration_ticker = time.NewTicker(time.Minute)
			go func() {
				for {
					_, ok := <-self.expiration_ticker.C
					if ok {
						t := time.Now()
						self.expire.ExpireBefore(t.Add(since))
					} else {
						return
					}
				}
			}()
		}
	}
	// we are now running
	self.running = true
	// start polling feeds
	go self.pollfeeds()
	threads := 8
	go func() {
		// if we have no initial posts create one
		if self.database.ArticleCount() == 0 {
			nntp := newPlaintextArticle("welcome to nntpchan, this post was inserted on startup automatically", "system@"+self.instance_name, "Welcome to NNTPChan", "system", self.instance_name, genMessageID(self.instance_name), "overchan.test")
			nntp.Pack()
			file := self.store.CreateFile(nntp.MessageID())
			if file != nil {
				err = nntp.WriteTo(file, MaxMessageSize)
				file.Close()
				if err == nil {
					self.loadFromInfeed(nntp.MessageID())
					nntp.Reset()
				} else {
					log.Println("failed to create startup messge?", err)
				}
			}
		}
	}()

	// get all pending articles from infeed and load them
	go func() {
		f, err := os.Open(self.store.TempDir())
		if err == nil {
			names, err := f.Readdirnames(0)
			if err == nil {
				for _, name := range names {
					self.loadFromInfeed(name)
				}
			}
		}

	}()
	// register feeds from config
	log.Println("registering feeds")
	for _, f := range self.conf.feeds {
		self.register_feed <- f
	}

	for threads > 0 {
		// fork off N go routines for handling messages
		go self.poll(threads)
		log.Println("started worker", threads)
		threads--
	}
	// start accepting incoming connections
	self.acceptloop()
	<-self.done
	// clean up pidfile if it was specified
	if pidfile != "" {
		os.Remove(pidfile)
	}
}

func (self *NNTPDaemon) syncAllMessages() {
	log.Println("syncing all messages to all feeds")
	for _, article := range self.database.GetAllArticles() {
		if self.store.HasArticle(article.MessageID()) {
			self.sendAllFeeds(article)
		}
	}
	log.Println("sync all messages queue flushed")
}

// load a message from the infeed directory
func (self *NNTPDaemon) loadFromInfeed(msgid string) {
	self.store.AcceptTempArticle(msgid)
	go self.processMessage(msgid)
}

// reload all configs etc
func (self *NNTPDaemon) Reload() {
	log.Println("reload daemon")
	conf := ReadConfig()
	if conf == nil {
		log.Println("failed to reload config")
		return
	}
	script, ok := conf.frontend["markup_script"]
	if ok {
		err := SetMarkupScriptFile(script)
		if err != nil {
			log.Println("failed to reload script file", err)
		}
	}
	log.Println("reload daemon okay")
}

func (self *NNTPDaemon) pollfeeds() {

	for {
		select {

		case q := <-self.get_feed:
			// someone asked for the status of a certain feed
			name := q.name
			// find feed
			feedstate, ok := self.loadedFeeds[name]
			if ok {
				// it exists
				if q.resultChnl != nil {
					// caller wants to be informed
					// create the reply
					status := &feedStatus{
						Exists: true,
						State:  feedstate,
					}
					// get the connections for this feed
					for _, conn := range self.activeConnections {
						if conn.feedname == name {
							status.Conns = append(status.Conns, conn)
						}
					}
					// tell caller
					q.resultChnl <- status
				}
			} else {
				// does not exist
				if q.resultChnl != nil {
					// tell caller
					q.resultChnl <- &feedStatus{
						Exists: false,
					}
				}
			}
		case ev := <-self.modify_feed_policy:
			// we want to modify a feed policy
			name := ev.name
			// does this feed exist?
			feedstate, ok := self.loadedFeeds[name]
			if ok {
				// yeh
				// replace the policy
				feedstate.Config.policy = ev.policy
				if ev.resultChnl != nil {
					// we need to inform the caller about the feed being changed successfully
					ev.resultChnl <- &modifyFeedPolicyResult{
						err:  nil,
						name: name,
					}
				}
			} else {
				// nah
				if ev.resultChnl != nil {
					// we need to inform the caller about the feed not existing
					ev.resultChnl <- &modifyFeedPolicyResult{
						err:  errors.New("no such feed"),
						name: name,
					}
				}
			}
		case chnl := <-self.get_feeds:
			// we got a request for viewing the status of the feeds
			var feeds []*feedStatus
			for feedname, feedstate := range self.loadedFeeds {
				var conns []*nntpConnection
				// get connections for this feed
				for _, conn := range self.activeConnections {
					if conn.feedname == feedname {
						conns = append(conns, conn)
					}
				}
				// add feedStatus
				feeds = append(feeds, &feedStatus{
					Exists: true,
					Conns:  conns,
					State:  feedstate,
				})
			}
			// send response
			chnl <- feeds
		case feedconfig := <-self.register_feed:
			self.loadedFeeds[feedconfig.Name] = &feedState{
				Config: &feedconfig,
			}
			log.Println("daemon registered feed", feedconfig.Name)
			// persist feeds
			if feedconfig.sync {
				go self.persistFeed(&feedconfig, "sync", 0)
			}
			n := feedconfig.connections
			for n > 0 {
				go self.persistFeed(&feedconfig, "stream", n)
				go self.persistFeed(&feedconfig, "reader", n)
				n--
			}
		case feedname := <-self.deregister_feed:
			_, ok := self.loadedFeeds[feedname]
			if ok {
				delete(self.loadedFeeds, feedname)
				log.Println("daemon deregistered feed", feedname)
			} else {
				log.Println("daemon does not have registered feed", feedname)
			}
		case outfeed := <-self.register_connection:
			self.activeConnections[outfeed.name] = outfeed
		case outfeed := <-self.deregister_connection:
			delete(self.activeConnections, outfeed.name)
		case <-self.pump_ticker.C:
			go self.pump_article_requests()
		}
	}
}

func (self *NNTPDaemon) informHooks(group, msgid, ref string) {
	if ValidMessageID(msgid) && ValidMessageID(ref) && ValidNewsgroup(group) {
		for _, conf := range self.conf.hooks {
			conf.Exec(group, msgid, ref)
		}
	}
}

func (self *NNTPDaemon) pump_article_requests() {
	var articles []ArticleEntry
	self.send_articles_mtx.Lock()
	articles = append(articles, self.send_articles...)
	self.send_articles = nil
	self.send_articles_mtx.Unlock()
	for _, entry := range articles {
		self.send_all_feeds <- entry
	}
	articles = nil
	self.ask_articles_mtx.Lock()
	msgids := self.ask_articles
	self.ask_articles = nil
	self.ask_articles_mtx.Unlock()
	for _, entry := range msgids {
		self.ask_for_article <- entry
	}
	articles = nil
}

func (self *NNTPDaemon) processMessage(msgid string) {
	log.Println("load", msgid)
	hdr := self.store.GetHeaders(msgid)
	if hdr == nil {
		log.Println("failed to load", msgid)
	} else {
		rollover := 100
		group := hdr.Get("Newsgroups", "")
		ref := hdr.Get("References", "")
		log.Println("got", msgid, "in", group, "references", ref != "")
		tpp, err := self.database.GetThreadsPerPage(group)
		ppb, err := self.database.GetPagesPerBoard(group)
		if err == nil {
			rollover = tpp * ppb
		}
		if self.expire != nil {
			// expire posts
			log.Println("expire", group, "for", rollover, "threads")
			go self.expire.ExpireGroup(group, rollover)
		}
		// send to mod panel
		if group == "ctl" {
			log.Println("process mod message", msgid)
			go self.mod.HandleMessage(msgid)
		}
		// inform callback hooks
		go self.informHooks(group, msgid, ref)
		// federate
		go self.sendAllFeeds(ArticleEntry{msgid, group})
		// send to frontend
		if self.frontend != nil {
			if self.frontend.AllowNewsgroup(group) {
				self.frontend.HandleNewPost(frontendPost{msgid, ref, group})
			}
		}
	}

}

func (self *NNTPDaemon) poll(worker int) {
	for {
		select {
		case nntp := <-self.send_all_feeds:
			group := nntp.Newsgroup()
			if self.Federate() {
				sz, _ := self.store.GetMessageSize(nntp.MessageID())
				feeds := self.activeFeeds()
				if feeds != nil {
					for _, f := range feeds {
						var send []*nntpConnection
						for _, feed := range f.Conns {
							if feed.policy.AllowsNewsgroup(group) {
								if strings.HasSuffix(feed.name, "-stream") {
									send = append(send, feed)
								}
							}
						}
						minconn := lowestBacklogConnection(send)
						if minconn != nil {
							go minconn.offerStream(nntp.MessageID(), sz)
						}
					}
				}
			}
		case nntp := <-self.ask_for_article:
			feeds := self.activeFeeds()
			if feeds != nil {
				for _, f := range feeds {
					var send []*nntpConnection
					for _, feed := range f.Conns {
						if strings.HasSuffix(feed.name, "-reader") {
							send = append(send, feed)
						}
					}
					minconn := lowestBacklogConnection(send)
					if minconn != nil {
						go minconn.askForArticle(nntp)
					}
				}
			}
		}
	}
	log.Println("worker", worker, "done")
}

// get connection with smallest backlog
func lowestBacklogConnection(conns []*nntpConnection) (minconn *nntpConnection) {
	min := int64(0)
	for _, c := range conns {
		b := c.GetBacklog()
		if min == 0 || b < min {
			minconn = c
			min = b
		}
	}
	return
}

func (self *NNTPDaemon) askForArticle(msgid string) {
	self.ask_articles_mtx.Lock()
	self.ask_articles = append(self.ask_articles, msgid)
	self.ask_articles_mtx.Unlock()
}

func (self *NNTPDaemon) sendAllFeeds(e ArticleEntry) {
	self.send_articles_mtx.Lock()
	self.send_articles = append(self.send_articles, e)
	self.send_articles_mtx.Unlock()
}

func (self *NNTPDaemon) acceptloop() {
	for {
		// accept
		conn, err := self.listener.Accept()
		if err != nil {
			log.Fatal(err)
		}
		// make a new inbound nntp connection handler
		hostname := ""
		if self.conf.crypto != nil {
			hostname = self.conf.crypto.hostname
		}
		nntp := createNNTPConnection(hostname)
		if self.conf.daemon["anon_nntp"] == "1" {
			nntp.authenticated = true
		}
		addr := conn.RemoteAddr()
		nntp.name = fmt.Sprintf("%s-inbound-feed", addr.String())
		nntp.policy = self.conf.inboundPolicy
		c := textproto.NewConn(conn)
		// send banners and shit
		err = nntp.inboundHandshake(c)
		if err == nil {
			// run, we support stream and reader
			go nntp.runConnection(self, true, true, true, false, "stream", conn, nil)
		} else {
			log.Println("failed to send banners", err)
			c.Close()
		}
	}
}

func (self *NNTPDaemon) Federate() (federate bool) {
	federate = len(self.conf.feeds) > 0
	return
}

func (self *NNTPDaemon) GetOurTLSConfig() *tls.Config {
	return self.GetTLSConfig(self.conf.crypto.hostname)
}

func (self *NNTPDaemon) GetTLSConfig(hostname string) *tls.Config {
	cfg := self.tls_config
	return &tls.Config{
		ServerName:   hostname,
		CipherSuites: cfg.CipherSuites,
		RootCAs:      cfg.RootCAs,
		ClientCAs:    cfg.ClientCAs,
		Certificates: cfg.Certificates,
		ClientAuth:   cfg.ClientAuth,
	}
}

func (self *NNTPDaemon) RequireTLS() (require bool) {
	v, ok := self.conf.daemon["require_tls"]
	if ok {
		require = v == "1"
	}
	return
}

// return true if we can do tls
func (self *NNTPDaemon) CanTLS() (can bool) {
	can = self.tls_config != nil
	return
}

func (self *NNTPDaemon) Setup() {
	log.Println("checking for configs...")
	// check that are configs exist
	CheckConfig()
	log.Println("loading config...")
	// read the config
	self.conf = ReadConfig()
	if self.conf == nil {
		log.Fatal("failed to load config")
	}
	// validate the config
	log.Println("validating configs...")
	self.conf.Validate()
	log.Println("configs are valid")

	var err error

	log.Println("Reading translation files")
	translation_dir := self.conf.frontend["translations"]
	if translation_dir == "" {
		translation_dir = filepath.Join("contrib", "translations")
	}
	locale := self.conf.frontend["locale"]
	InitI18n(locale, translation_dir)

	db_host := self.conf.database["host"]
	db_port := self.conf.database["port"]
	db_user := self.conf.database["user"]
	db_passwd := self.conf.database["password"]

	var ok bool
	var val string

	// set up database stuff
	log.Println("connecting to database...")
	self.database = NewDatabase(self.conf.database["type"], self.conf.database["schema"], db_host, db_port, db_user, db_passwd)
	if val, ok = self.conf.database["connidle"]; ok {
		i, _ := strconv.Atoi(val)
		if i > 0 {
			self.database.SetMaxIdleConns(i)
		}
	}
	if val, ok = self.conf.database["maxconns"]; ok {
		i, _ := strconv.Atoi(val)
		if i > 0 {
			self.database.SetMaxOpenConns(i)
		}
	}
	if val, ok = self.conf.database["connlife"]; ok {
		i, _ := strconv.Atoi(val)
		if i > 0 {
			self.database.SetConnectionLifetime(i)
		}
	}
	log.Println("ensure that the database is created...")
	self.database.CreateTables()

	// ensure tls stuff
	if self.conf.crypto != nil {
		self.tls_config, err = GenTLS(self.conf.crypto)
		if err != nil {
			log.Fatal("failed to initialize tls: ", err)
		}
	}

	// set up store
	log.Println("set up article store...")
	self.store = createArticleStore(self.conf.store, self.conf.thumbnails, self.database, &self.spamFilter)

	// do we enable the frontend?
	if self.conf.frontend["enable"] == "1" {
		log.Printf("frontend %s enabled", self.conf.frontend["name"])

		cache_host := self.conf.cache["host"]
		cache_port := self.conf.cache["port"]
		cache_user := self.conf.cache["user"]
		cache_passwd := self.conf.cache["password"]
		self.cache = NewCache(self.conf.cache["type"], cache_host, cache_port, cache_user, cache_passwd, self.conf.cache, self.conf.frontend, self.database, self.store)

		script, ok := self.conf.frontend["markup_script"]
		if ok {
			err = SetMarkupScriptFile(script)
			if err != nil {
				log.Println("failed to load markup script", err)
			}
		}

		self.frontend = NewHTTPFrontend(self, self.cache, self.conf.frontend, self.conf.worker["url"])
	}

	self.spamFilter.Configure(self.conf.spamconf)

	regen := func(string, string, string, int) {}
	if self.frontend != nil {
		regen = self.frontend.RegenOnModEvent
	}
	self.mod = &modEngine{
		//spam:     &self.spamFilter,
		store:    self.store,
		database: self.database,
		regen:    regen,
	}
	// inject DB into template engine
	template.DB = self.database
}

func (daemon *NNTPDaemon) ModEngine() ModEngine {
	return daemon.mod
}
