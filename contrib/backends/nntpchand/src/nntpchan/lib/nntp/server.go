package nntp

import (
	log "github.com/Sirupsen/logrus"
	"net"
	"nntpchan/lib/config"
	"nntpchan/lib/network"
	"nntpchan/lib/store"
	"time"
)

// nntp outfeed state
type nntpFeed struct {
	conn Conn
	send chan ArticleEntry
	conf *config.FeedConfig
}

// an nntp server
type Server struct {
	// user callback
	Hooks EventHooks
	// filters to apply
	Filters []ArticleFilter
	// global article acceptor
	Acceptor ArticleAcceptor
	// article storage
	Storage store.Storage
	// nntp config
	Config *config.NNTPServerConfig
	// outfeeds to connect to
	Feeds []*config.FeedConfig
	// inbound authentiaction mechanism
	Auth ServerAuth
	// send to outbound feed channel
	send chan ArticleEntry
	// register inbound feed channel
	regis chan *nntpFeed
	// deregister inbound feed channel
	deregis chan *nntpFeed
}

func NewServer() *Server {
	return &Server{
		// XXX: buffered?
		send:    make(chan ArticleEntry),
		regis:   make(chan *nntpFeed),
		deregis: make(chan *nntpFeed),
	}
}

// reload server configuration
func (s *Server) ReloadServer(c *config.NNTPServerConfig) {

}

// reload feeds
func (s *Server) ReloadFeeds(feeds []*config.FeedConfig) {

}

func (s *Server) GotArticle(msgid MessageID, group Newsgroup) {
	log.WithFields(log.Fields{
		"pkg":   "nntp-server",
		"msgid": msgid,
		"group": group,
	}).Info("obtained article")
	if s.Hooks != nil {
		s.Hooks.GotArticle(msgid, group)
	}
	// send to outbound feeds
	s.send <- ArticleEntry{msgid.String(), group.String()}
}

func (s *Server) SentArticleVia(msgid MessageID, feedname string) {
	log.WithFields(log.Fields{
		"pkg":   "nntp-server",
		"msgid": msgid,
		"feed":  feedname,
	}).Info("article sent")
	if s.Hooks != nil {
		s.Hooks.SentArticleVia(msgid, feedname)
	}
}

func (s *Server) Name() string {
	if s.Config == nil || s.Config.Name == "" {
		return "nntp.anon.tld"
	}
	return s.Config.Name
}

// persist 1 feed forever
func (s *Server) persist(cfg *config.FeedConfig) {
	delay := time.Second

	log.WithFields(log.Fields{
		"name": cfg.Name,
	}).Debug("Persist Feed")
	for {
		dialer := network.NewDialer(cfg.Proxy)
		c, err := dialer.Dial(cfg.Addr)
		if err == nil {
			// successful connect
			delay = time.Second
			conn := newOutboundConn(c, s, cfg)
			err = conn.Negotiate(true)
			if err == nil {
				// negotiation good
				log.WithFields(log.Fields{
					"name": cfg.Name,
				}).Debug("Negotitation good")
				// start streaming
				var chnl chan ArticleEntry
				chnl, err = conn.StartStreaming()
				if err == nil {
					// register new connection
					f := &nntpFeed{
						conn: conn,
						send: chnl,
						conf: cfg,
					}
					s.regis <- f
					// start streaming
					conn.StreamAndQuit()
					// deregister
					s.deregis <- f
					continue
				}
			} else {
				log.WithFields(log.Fields{
					"name": cfg.Name,
				}).Info("outbound nntp connection failed to negotiate ", err)
			}
			conn.Quit()
		} else {
			// failed dial, do exponential backoff up to 1 hour
			if delay <= time.Hour {
				delay *= 2
			}
			log.WithFields(log.Fields{
				"name": cfg.Name,
			}).Info("feed backoff for ", delay)
			time.Sleep(delay)
		}
	}
}

// download all new posts from a remote server
func (s *Server) downloadPosts(cfg *config.FeedConfig) error {
	dialer := network.NewDialer(cfg.Proxy)
	c, err := dialer.Dial(cfg.Addr)
	if err != nil {
		return err
	}
	conn := newOutboundConn(c, s, cfg)
	err = conn.Negotiate(false)
	if err != nil {
		conn.Quit()
		return err
	}
	groups, err := conn.ListNewsgroups()
	if err != nil {
		conn.Quit()
		return err
	}
	for _, g := range groups {
		if cfg.Policy != nil && cfg.Policy.AllowGroup(g.String()) {
			log.WithFields(log.Fields{
				"group": g,
				"pkg":   "nntp-server",
			}).Debug("downloading group")
			err = conn.DownloadGroup(g)
			if err != nil {
				conn.Quit()
				return err
			}
		}
	}
	conn.Quit()
	return nil
}

func (s *Server) periodicDownload(cfg *config.FeedConfig) {
	for cfg.PullInterval > 0 {
		err := s.downloadPosts(cfg)
		if err != nil {
			// report error
			log.WithFields(log.Fields{
				"feed":  cfg.Name,
				"pkg":   "nntp-server",
				"error": err,
			}).Error("periodic download failed")
		}
		time.Sleep(time.Minute * time.Duration(cfg.PullInterval))
	}
}

// persist all outbound feeds
func (s *Server) PersistFeeds() {
	for _, f := range s.Feeds {
		go s.persist(f)
		go s.periodicDownload(f)
	}

	feeds := make(map[string]*nntpFeed)

	for {
		select {
		case e, ok := <-s.send:
			if !ok {
				break
			}
			msgid := e.MessageID().String()
			group := e.Newsgroup().String()
			// TODO: determine anon
			anon := false
			// TODO: determine attachments
			attachments := false

			for _, f := range feeds {
				if f.conf.Policy != nil && !f.conf.Policy.Allow(msgid, group, anon, attachments) {
					// not allowed in this feed
					continue
				}
				log.WithFields(log.Fields{
					"name":  f.conf.Name,
					"msgid": msgid,
					"group": group,
				}).Debug("sending article")
				f.send <- e
			}
			break
		case f, ok := <-s.regis:
			if ok {
				log.WithFields(log.Fields{
					"name": f.conf.Name,
				}).Debug("register feed")
				feeds[f.conf.Name] = f
			}
			break
		case f, ok := <-s.deregis:
			if ok {
				log.WithFields(log.Fields{
					"name": f.conf.Name,
				}).Debug("deregister feed")
				delete(feeds, f.conf.Name)
			}
			break
		}
	}
}

// serve connections from listener
func (s *Server) Serve(l net.Listener) (err error) {
	log.WithFields(log.Fields{
		"pkg":  "nntp-server",
		"addr": l.Addr(),
	}).Debug("Serving")
	for err == nil {
		var c net.Conn
		c, err = l.Accept()
		if err == nil {
			// we got a new connection
			go s.handleInboundConnection(c)
		} else {
			log.WithFields(log.Fields{
				"pkg": "nntp-server",
			}).Error("failed to accept inbound connection", err)
		}
	}
	return
}

// get the article policy for a connection given its state
func (s *Server) getPolicyFor(state *ConnState) ArticleAcceptor {
	return s.Acceptor
}

// recv inbound streaming messages
func (s *Server) recvInboundStream(chnl chan ArticleEntry) {
	for {
		e, ok := <-chnl
		if ok {
			s.GotArticle(e.MessageID(), e.Newsgroup())
		} else {
			return
		}
	}
}

// process an inbound connection
func (s *Server) handleInboundConnection(c net.Conn) {
	log.WithFields(log.Fields{
		"pkg":  "nntp-server",
		"addr": c.RemoteAddr(),
	}).Debug("handling inbound connection")
	var nc Conn
	nc = newInboundConn(s, c)
	err := nc.Negotiate(true)
	if err == nil {
		// do they want to stream?
		if nc.WantsStreaming() {
			// yeeeeeh let's stream
			var chnl chan ArticleEntry
			chnl, err = nc.StartStreaming()
			// for inbound we will recv messages
			go s.recvInboundStream(chnl)
			nc.StreamAndQuit()
			log.WithFields(log.Fields{
				"pkg":  "nntp-server",
				"addr": c.RemoteAddr(),
			}).Info("streaming finished")
			return
		} else {
			// handle non streaming commands
			nc.ProcessInbound(s)
		}
	} else {
		log.WithFields(log.Fields{
			"pkg":  "nntp-server",
			"addr": c.RemoteAddr(),
		}).Warn("failed to negotiate with inbound connection", err)
		c.Close()
	}
}
