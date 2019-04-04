//
// nntp.go -- nntp interface for peering
//
package srnd

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/mail"
	"net/textproto"
	"strconv"
	"strings"
	"sync"
	"time"
)

const nntpDummyArticle = "<keepalive@dummy.tld>"

type nntpStreamEvent string

func (ev nntpStreamEvent) MessageID() string {
	return strings.Split(string(ev), " ")[1]
}

func (ev nntpStreamEvent) Command() string {
	return strings.Split(string(ev), " ")[0]
}

func nntpTAKETHIS(msgid string) nntpStreamEvent {
	return nntpStreamEvent(fmt.Sprintf("TAKETHIS %s", msgid))
}

func nntpCHECK(msgid string) nntpStreamEvent {
	return nntpStreamEvent(fmt.Sprintf("CHECK %s", msgid))
}

type syncEvent struct {
	msgid string
	sz    int64
	state string
}

// nntp connection state
type nntpConnection struct {
	// the name of the feed this connection belongs to
	feedname string
	// the name of this connection
	name string
	// hostname used for tls
	hostname string
	// the mode we are in now
	mode string
	// what newsgroup is currently selected or empty string if none is selected
	group string
	// what article is currently selected
	selected_article string
	// the policy for federation
	policy *FeedPolicy
	// lock help when expecting non pipelined activity
	access sync.Mutex

	// pending articles to request
	articles []string
	// lock for accessing articles to request
	articles_access sync.Mutex
	// map of message-id -> stream state
	pending map[string]*syncEvent
	// lock for accessing self.pending map
	pending_access sync.Mutex

	tls_state tls.ConnectionState

	// have we authenticated with a login?
	authenticated bool
	// the username that is authenticated
	username string
	// send a channel down this channel to be informed when streaming/reader dies when commanded by QuitAndWait()
	die chan chan bool
	// remote address of this connections
	addr net.Addr
	// pending backlog of bytes to transfer
	backlog int64
	// function used to close connection abruptly
	abort func()

	// streaming keepalive timer
	keepalive *time.Ticker
}

// get message backlog in bytes
func (self *nntpConnection) GetBacklog() int64 {
	return self.backlog
}

func (self *nntpConnection) MarshalJSON() (data []byte, err error) {
	jmap := make(map[string]interface{})
	pending := make(map[string]string)
	self.pending_access.Lock()
	for k, v := range self.pending {
		pending[k] = v.state
	}
	self.pending_access.Unlock()
	jmap["pending"] = pending
	jmap["mode"] = self.mode
	jmap["name"] = self.name
	jmap["authed"] = self.authenticated
	jmap["group"] = self.group
	jmap["backlog"] = self.backlog
	data, err = json.Marshal(jmap)
	return
}

func createNNTPConnection(addr string) *nntpConnection {
	var host string
	if len(addr) > 0 {
		host, _, _ = net.SplitHostPort(addr)
	}
	return &nntpConnection{
		hostname: host,
		pending:  make(map[string]*syncEvent),
	}
}

// gracefully exit nntpconnection when all active transfers are done
// returns when that is done
func (self *nntpConnection) QuitAndWait() {
	chnl := make(chan bool)
	// tell feed to die
	self.die <- chnl
	// we are ded
	<-chnl
	// ded now
	close(chnl)
	return
}

// switch modes
func (self *nntpConnection) modeSwitch(mode string, conn *textproto.Conn) (success bool, err error) {
	self.access.Lock()
	mode = strings.ToUpper(mode)
	err = conn.PrintfLine("MODE %s", mode)
	if err != nil {
		log.Println(self.name, "cannot switch mode", err)
		self.access.Unlock()
		return
	}
	var code int
	code, _, err = conn.ReadCodeLine(-1)
	if code >= 200 && code < 300 {
		// accepted mode change
		if len(self.mode) > 0 {
			log.Printf(self.name, "mode switch %s -> %s", self.mode, mode)
		} else {
			log.Println(self.name, "switched to mode", mode)
		}
		self.mode = mode
		success = len(self.mode) > 0
	}
	self.access.Unlock()
	return
}

func (self *nntpConnection) Quit(conn *textproto.Conn) (err error) {
	conn.PrintfLine("QUIT")
	_, _, err = conn.ReadCodeLine(0)
	return
}

// send a banner for inbound connections
func (self *nntpConnection) inboundHandshake(conn *textproto.Conn) (err error) {
	err = conn.PrintfLine("200 Posting Allowed")
	return err
}

// outbound setup, check capabilities and set mode
// returns (supports stream, supports reader, supports tls) + error
func (self *nntpConnection) outboundHandshake(conn *textproto.Conn, conf *FeedConfig) (stream, reader, tls bool, err error) {
	log.Println(self.name, "outbound handshake")
	var line string
	var code int
	for err == nil {
		code, line, err = conn.ReadCodeLine(-1)
		log.Println(self.name, line)
		if err == nil {
			if code == 200 {
				// send capabilities
				log.Println(self.name, "ask for capabilities")
				err = conn.PrintfLine("CAPABILITIES")
				if err == nil {
					// read response
					dr := conn.DotReader()
					r := bufio.NewReader(dr)
					for {
						line, err = r.ReadString('\n')
						if err == io.EOF {
							// we are at the end of the dotreader
							// set err back to nil and break out
							err = nil
							break
						} else if err == nil {
							if line == "STARTTLS\n" {
								log.Println(self.name, "supports STARTTLS")
								tls = true
							} else if line == "MODE-READER\n" || line == "READER\n" {
								log.Println(self.name, "supports READER")
								reader = true
							} else if line == "STREAMING\n" {
								stream = true
								log.Println(self.name, "supports STREAMING")
							} else if line == "POSTIHAVESTREAMING\n" {
								stream = true
								reader = false
								log.Println(self.name, "is SRNd")
							}
						} else {
							// we got an error
							log.Println("error reading capabilities", err)
							break
						}
					}
					// return after reading
					break
				}
			} else if code == 201 {
				log.Println("feed", self.name, "does not allow posting")
				// we don't do auth yet
				break
			} else {
				continue
			}
		}
	}
	if conf != nil && len(conf.username) > 0 && len(conf.passwd) > 0 {
		log.Println(self.name, "authenticating...")
		err = conn.PrintfLine("AUTHINFO USER %s", conf.username)
		if err == nil {
			var code int
			code, line, err = conn.ReadCodeLine(381)
			if code == 381 {
				err = conn.PrintfLine("AUTHINFO PASS %s", conf.passwd)
				if err == nil {
					code, line, err = conn.ReadCodeLine(281)
					if code == 281 {
						log.Println(self.name, "Auth Successful")
						self.authenticated = true
					} else {
						log.Println(self.name, "Auth incorrect", line)
						conn.PrintfLine("QUIT")
						conn.Close()
						return false, false, false, io.EOF
					}
				}
			}
		}
	}

	return
}

// offer up a article to sync via this connection
func (self *nntpConnection) offerStream(msgid string, sz int64) {
	if self.messageIsQueued(msgid) {
		// already queued for send
	} else {
		self.backlog += sz
		self.messageSetPendingState(msgid, "CHECK", sz)
	}
}

// handle sending 1 stream event
func (self *nntpConnection) handleStreamEvent(ev nntpStreamEvent, daemon *NNTPDaemon, conn *textproto.Conn) (err error) {
	if ValidMessageID(ev.MessageID()) {
		cmd, msgid := ev.Command(), ev.MessageID()
		if cmd == "TAKETHIS" {
			// open message for reading
			var rc io.ReadCloser
			rc, err = daemon.store.OpenMessage(msgid)
			if err == nil {
				err = conn.PrintfLine("%s", ev)
				// time to send
				dw := conn.DotWriter()
				_, err = io.Copy(dw, rc)
				rc.Close()
				err = dw.Close()
				self.backlog -= self.pending[msgid].sz
				delete(self.pending, msgid)
			} else {
				self.backlog -= self.pending[msgid].sz
				delete(self.pending, msgid)
				// ignore this error
				err = nil
			}
		} else if cmd == "CHECK" {
			conn.PrintfLine("%s", ev)
			self.pending[msgid].state = "pending"
		} else {
			log.Println("invalid stream command", ev)
		}
	}
	return
}

// get all articles give their streaming state
func (self *nntpConnection) getArticlesInState(state string) (articles []string) {
	self.pending_access.Lock()
	for _, ev := range self.pending {
		if ev.state == state {
			articles = append(articles, ev.msgid)
		}
	}
	self.pending_access.Unlock()
	return
}

func (self *nntpConnection) messageIsQueued(msgid string) (queued bool) {
	self.pending_access.Lock()
	_, queued = self.pending[msgid]
	self.pending_access.Unlock()
	return
}

func (self *nntpConnection) messageSetPendingState(msgid, state string, sz int64) {
	self.pending_access.Lock()
	s, has := self.pending[msgid]
	if has {
		s.state = state
		self.pending[msgid] = s
	} else {
		self.pending[msgid] = &syncEvent{msgid: msgid, sz: sz, state: state}
	}
	self.pending_access.Unlock()
}

func (self *nntpConnection) messageSetProcessed(msgid string) {
	self.pending_access.Lock()
	ev, ok := self.pending[msgid]
	if ok {
		self.backlog -= ev.sz
		delete(self.pending, msgid)
	}
	self.pending_access.Unlock()
}

// handle streaming events
// this function should send only
func (self *nntpConnection) handleStreaming(daemon *NNTPDaemon, conn *textproto.Conn) (err error) {
	for err == nil {
		select {
		case chnl := <-self.die:
			// someone asked us to die
			self.pending_access.Lock()
			conn.PrintfLine("QUIT")
			conn.Close()
			self.pending_access.Unlock()
			chnl <- true
			return
		case <-self.keepalive.C:
			self.pending_access.Lock()
			err = conn.PrintfLine("CHECK %s", nntpDummyArticle)
			self.pending_access.Unlock()
		default:
			if len(self.pending) > 0 {
				self.pending_access.Lock()
				for msgid, ev := range self.pending {
					if ev.state == "CHECK" {
						err = self.handleStreamEvent(nntpCHECK(msgid), daemon, conn)
					} else if ev.state == "TAKETHIS" {
						err = self.handleStreamEvent(nntpTAKETHIS(msgid), daemon, conn)
					}
				}
				self.pending_access.Unlock()
			}
			time.Sleep(time.Millisecond * 100)
		}
	}
	return
}

// check if we want the article given its auth status and mime header
// returns empty string if it's okay otherwise an error message
func (self *nntpConnection) checkMIMEHeader(daemon *NNTPDaemon, hdr textproto.MIMEHeader) (reason string, allow bool, err error) {
	if !self.authenticated {
		reason = "not authenticated"
		return
	}
	reason, allow, err = self.checkMIMEHeaderNoAuth(daemon, hdr)
	return
}

// check if we want the article given its mime header without checking auth status
// returns empty string if it's okay otherwise an error message
func (self *nntpConnection) checkMIMEHeaderNoAuth(daemon *NNTPDaemon, hdr textproto.MIMEHeader) (reason string, ban bool, err error) {
	newsgroup := hdr.Get("Newsgroups")
	reference := hdr.Get("References")
	msgid := getMessageID(hdr)
	encaddr := hdr.Get("X-Encrypted-Ip")
	torposter := hdr.Get("X-Tor-Poster")
	i2paddr := hdr.Get("X-I2p-Desthash")
	content_type := hdr.Get("Content-Type")
	has_attachment := strings.HasPrefix(content_type, "multipart/mixed")
	pubkey := hdr.Get("X-Pubkey-Ed25519")
	// TODO: allow certain pubkeys?
	is_signed := pubkey != ""
	is_ctl := newsgroup == "ctl" && is_signed
	anon_poster := torposter != "" || i2paddr != "" || encaddr == ""

	server_pubkey := hdr.Get("X-Frontend-Pubkey")
	server_sig := hdr.Get("X-Frontend-Signature")

	is_spam := strings.HasPrefix(hdr.Get("X-Spam-Status"), "Yes,")

	if is_spam {
		reason = "message marked as spam by SpamAssassin"
		ban = true
		return
	}

	if serverPubkeyIsValid(server_pubkey) {
		b, _ := daemon.database.PubkeyRejected(server_pubkey)
		if b {
			reason = "server's pubkey is banned"
			ban = true
			return
		}
		if !verifyFrontendSig(server_pubkey, server_sig, msgid) {
			reason = "invalid frontend signature"
			ban = true
			return
		}
	} else if server_pubkey != "" {
		reason = fmt.Sprintf("invalid server public key: %s", server_pubkey)
		ban = true
		return
	}

	if !newsgroupValidFormat(newsgroup) {
		// invalid newsgroup format
		reason = fmt.Sprintf("invalid newsgroup: %s", newsgroup)
		ban = true
		return
	} else if banned, _ := daemon.database.NewsgroupBanned(newsgroup); banned {
		reason = "newsgroup banned"
		ban = true
		return
	} else if banned, _ = daemon.database.PubkeyRejected(pubkey); banned {
		// check for banned pubkey
		reason = "poster's pubkey is banned"
		ban = true
		return
	} else if self.policy != nil && !self.policy.AllowsNewsgroup(newsgroup) {
		reason = "newsgroup not allowed by feed policy"
		ban = true
		return
	} else if !(ValidMessageID(msgid) || (reference != "" && !ValidMessageID(reference))) {
		// invalid message id or reference
		reason = "invalid reference or message id is '" + msgid + "' reference is '" + reference + "'"
		ban = true
		return
	} else if daemon.store.HasArticle(msgid) {
		// we have already obtain this article locally
		reason = "we have this article locally"
		// don't ban
		return
	} else if daemon.database.ArticleBanned(msgid) {
		reason = "article banned"
		ban = true
		return
	} else if reference != "" && daemon.database.ArticleBanned(reference) {
		reason = "thread banned"
		ban = true
		return
	} else if daemon.database.HasArticle(msgid) {
		// this article is too old
		reason = "article already seen"
		// don't ban
		return
	} else if is_ctl {
		// we always allow control messages
		return
	} else if anon_poster {
		// this was posted anonymously
		if daemon.allow_anon {
			if has_attachment {
				// this has attachment
				if daemon.allow_anon_attachments {
					if daemon.allow_attachments {
						// we'll allow anon attachments
						return
					} else {
						// no attachments permitted
						reason = "no attachments allowed"
						ban = true
						return
					}
				} else {
					// we don't take signed messages or attachments posted anonymously
					reason = "no anon attachments"
					ban = true
					return
				}
			} else {
				// we allow anon posts that are plain
				return
			}
		} else {
			// we don't allow anon posts of any kind
			reason = "no anon posts allowed"
			ban = true
			return
		}
	} else {
		// check for banned address
		if encaddr != "" {
			ban, err = daemon.database.CheckEncIPBanned(encaddr)
			if err == nil {
				if ban {
					// this address is banned
					reason = "poster remote address is banned"
					return
				} else {
					// not banned
				}
			}
		} else {
			// idk wtf
			log.Println(self.name, "wtf? invalid article")
		}
	}
	if !daemon.allow_attachments {
		// we don't want attachments
		if is_ctl {
			// ctl is fine
			return
		} else if is_signed {
			// may have an attachment, reject
			reason = "disallow signed posts because no attachments allowed"
			ban = true
		} else if has_attachment {
			// we have an attachment, reject
			reason = "attachments of any kind not allowed"
			ban = true
		}
	}
	return
}

func (self *nntpConnection) handleLine(daemon *NNTPDaemon, code int, line string, conn *textproto.Conn) (err error) {
	parts := strings.Split(line, " ")
	var msgid string
	if code == 0 && len(parts) > 1 {
		msgid = parts[1]
	} else {
		msgid = parts[0]
	}
	if code == 238 {
		if msgid == nntpDummyArticle {
			return
		}

		sz, _ := daemon.store.GetMessageSize(msgid)
		self.messageSetPendingState(msgid, "TAKETHIS", sz)
		return
	} else if code == 239 {
		// successful TAKETHIS
		log.Println(msgid, "sent via", self.name)
		self.messageSetProcessed(msgid)
		return
		// TODO: remember success
	} else if code == 431 {
		if msgid == nntpDummyArticle {
			return
		}
		// CHECK said we would like this article later
		self.messageSetProcessed(msgid)
	} else if code == 439 {
		if msgid == nntpDummyArticle {
			return
		}
		// TAKETHIS failed
		log.Println(msgid, "was not sent to", self.name, "denied:", line)
		self.messageSetProcessed(msgid)
		// TODO: remember denial
	} else if code == 438 {
		if msgid == nntpDummyArticle {
			return
		}
		// they don't want the article
		// TODO: remeber rejection
		self.messageSetProcessed(msgid)
	} else {
		// handle command
		parts := strings.Split(line, " ")
		if len(parts) > 1 {
			cmd := strings.ToUpper(parts[0])
			if cmd == "MODE" {
				mode := strings.ToUpper(parts[1])
				if mode == "READER" {
					// reader mode
					self.mode = "READER"
					log.Println(self.name, "switched to reader mode")
					if self.authenticated {
						conn.PrintfLine("200 Posting Permitted")
					} else {
						conn.PrintfLine("201 No posting Permitted")
					}
				} else if mode == "STREAM" && self.authenticated {
					// wut? we're already in streaming mode
					log.Println(self.name, "already in streaming mode")
					conn.PrintfLine("203 Streaming enabled brah")
				} else {
					// invalid
					log.Println(self.name, "got invalid mode request", parts[1])
					conn.PrintfLine("501 invalid mode variant:", parts[1])
				}
			} else if strings.HasPrefix(line, "QUIT") {
				// quit command
				conn.PrintfLine("205 bai")
				// close our connection and return
				conn.Close()
				return

			} else if cmd == "AUTHINFO" {
				if len(parts) > 1 {
					auth_cmd := strings.ToUpper(parts[1])
					if auth_cmd == "USER" {
						// first part
						self.username = parts[2]
						// next phase is PASS
						conn.PrintfLine("381 Password required")
					} else if auth_cmd == "PASS" {
						if len(self.username) == 0 {
							conn.PrintfLine("482 Authentication commands issued out of sequence")
						} else {
							// try login
							var valid bool
							valid, err = daemon.database.CheckNNTPUserExists(self.username)
							if valid {
								valid, err = daemon.database.CheckNNTPLogin(self.username, line[14:])
							}
							if valid {
								// valid login
								self.authenticated = true
								conn.PrintfLine("281 Authentication accepted")
							} else if err == nil {
								// invalid login
								conn.PrintfLine("481 Authentication rejected")
							} else {
								// there was an error
								// logit
								log.Println(self.name, "error while logging in as", self.username, err)
								conn.PrintfLine("501 error while logging in")
							}
						}
					}
				} else {
					// wut ?
					// wrong legnth of parametrs
				}
			} else if cmd == "CHECK" {
				// handle check command
				msgid := parts[1]
				if !self.authenticated {
					// if client cannot TAKETHIS it shouldn't be able to CHECK either
					conn.PrintfLine("480 You have not authenticated")
					return
				}
				// have we seen this article?
				if daemon.store.HasArticle(msgid) {
					// yeh don't want it
					conn.PrintfLine("438 %s", msgid)
				} else if daemon.database.ArticleBanned(msgid) {
					// it's banned we don't want it
					conn.PrintfLine("438 %s", msgid)
				} else {
					// yes we do want it and we don't have it
					conn.PrintfLine("238 %s", msgid)
				}
			} else if cmd == "TAKETHIS" {
				// handle takethis command
				r := bufio.NewReader(conn.DotReader())
				if !self.authenticated {
					// early reject without parsing incase client not allowed to post
					// send response first to allow client to stop sending
					// XXX what happens if our send queue is full and this blocks?
					// other side will probably fill up sending article to us
					// async receive processing would help there
					// but this situation has low likehood to happen
					// operating system/transport buffering will compensate
					// so leave it this way
					conn.PrintfLine("480 You have not authenticated")
					// discard whole article without looking at insides
					// it should be dot-terminated either way
					_, err = io.Copy(ioutil.Discard, r)
					return
				}
				// client is allowed to post
				var msg *mail.Message
				var reason string
				var ban bool
				// read the article header
				msg, err = readMIMEHeader(r)
				if err != nil {
					log.Println(self.name, "error reading mime header:", err)
					conn.PrintfLine("439 %s error reading mime header", msgid)
					// if reading header error'd, msg.Body wont be set
					_, err = io.Copy(ioutil.Discard, r)
					return
				}
				hdr := textproto.MIMEHeader(msg.Header)
				// check the header
				reason, ban, err = self.checkMIMEHeaderNoAuth(daemon, hdr)
				if len(reason) > 0 {
					// discard, we do not want
					log.Println(self.name, "rejected", msgid, reason)
					conn.PrintfLine("439 %s %s", msgid, reason)
					_, err = io.Copy(ioutil.Discard, msg.Body)
					if ban {
						err = daemon.database.BanArticle(msgid, reason)
					}
				} else if err == nil {
					// looks good to accept
					// check if we don't have the rootpost
					reference := hdr.Get("References")
					if reference != "" && ValidMessageID(reference) && !daemon.store.HasArticle(reference) && !daemon.database.IsExpired(reference) {
						log.Println(self.name, "got reply to", reference, "but we don't have it")
						go daemon.askForArticle(reference)
					}
					err = storeMessage(daemon, hdr, msg.Body)
					if err == nil {
						code = 239
						reason = "gotten"
					} else {
						code = 439
						reason = err.Error()
					}
					conn.PrintfLine("%d %s %s", code, msgid, reason)
				} else {
					// error?
					// discard, we do not want
					log.Println(self.name, "rejected", msgid, "unexpected error", err)
					conn.PrintfLine("439 %s unexpected error", msgid)
					_, err = io.Copy(ioutil.Discard, msg.Body)
					if ban {
						err = daemon.database.BanArticle(msgid, reason)
					}
				}
			} else if cmd == "ARTICLE" {
				if !ValidMessageID(msgid) {
					if len(self.group) > 0 {
						n, err := strconv.Atoi(msgid)
						if err == nil {
							msgid, err = daemon.database.GetMessageIDForNNTPID(self.group, int64(n))
						}
					}
				}
				if ValidMessageID(msgid) {
					if daemon.database.ArticleBanned(msgid) {
						// article banned
						conn.PrintfLine("439 %s article banned from server", msgid)
					} else if daemon.store.HasArticle(msgid) {
						// we have it yeh
						f, err := daemon.store.OpenMessage(msgid)
						if err == nil {
							conn.PrintfLine("220 %s", msgid)
							dw := conn.DotWriter()
							_, err = io.Copy(dw, f)
							dw.Close()
							f.Close()
						} else {
							// wtf?!
							conn.PrintfLine("503 idkwtf happened: %s", err.Error())
						}
					} else {
						// we dont got it (by msgid)
						conn.PrintfLine("430 %s", msgid)
					}
				} else {
					// we dont got it (by num)
					conn.PrintfLine("423 %s", msgid)
				}
			} else if cmd == "IHAVE" {
				if !self.authenticated {
					conn.PrintfLine("480 You have not authenticated")
				} else {
					// handle IHAVE command
					msgid := parts[1]
					if daemon.database.HasArticleLocal(msgid) || daemon.database.HasArticle(msgid) || daemon.database.ArticleBanned(msgid) {
						// we don't want it
						conn.PrintfLine("435 Article Not Wanted")
					} else {
						// gib we want
						conn.PrintfLine("335 Send it plz")
						r := bufio.NewReader(conn.DotReader())
						msg, err := readMIMEHeader(r)
						if err == nil {
							// check the header
							hdr := textproto.MIMEHeader(msg.Header)
							var reason string
							var ban bool
							reason, ban, err = self.checkMIMEHeader(daemon, hdr)
							if len(reason) > 0 {
								// discard, we do not want
								log.Println(self.name, "rejected", msgid, reason)
								_, err = io.Copy(ioutil.Discard, r)
								if ban {
									_ = daemon.database.BanArticle(msgid, reason)
								}
								conn.PrintfLine("437 Rejected do not send again bro")
							} else {
								// check if we don't have the rootpost
								reference := hdr.Get("References")
								newsgroup := hdr.Get("Newsgroups")
								if reference != "" && ValidMessageID(reference) && !daemon.store.HasArticle(reference) && !daemon.database.IsExpired(reference) {
									log.Println(self.name, "got reply to", reference, "but we don't have it")
									go daemon.askForArticle(reference)
								}
								body := &io.LimitedReader{
									R: r,
									N: daemon.messageSizeLimitFor(newsgroup),
								}
								err = storeMessage(daemon, hdr, body)
								if err == nil {
									conn.PrintfLine("235 We got it")
								} else {
									conn.PrintfLine("437 Transfer Failed %s", err.Error())
								}
							}
						} else {
							// error here
							conn.PrintfLine("436 Transfer failed: " + err.Error())
						}
					}
				}
			} else if cmd == "LISTGROUP" {
				// handle LISTGROUP
				var group string
				if len(parts) > 1 {
					// parameters
					group = parts[1]
				} else {
					group = self.group
				}
				if len(group) > 0 && newsgroupValidFormat(group) {
					if daemon.database.HasNewsgroup(group) {
						// we has newsgroup
						var hi, lo int64
						// THIS is heavy as shit
						//count, err := daemon.database.CountAllArticlesInGroup(group)
						var err error
						count := 0
						if err == nil {
							hi, lo, err = daemon.database.GetLastAndFirstForGroup(group)
							if err == nil {
								conn.PrintfLine("211 %d %d %d %s list follows", count, lo, hi, group)
								dw := conn.DotWriter()
								idx := lo
								for idx <= hi {
									fmt.Fprintf(dw, "%d\r\n", idx)
									idx++
								}
								dw.Close()
							}
						}
						if err != nil {
							log.Println("LISTGROUP fail", err)
							conn.PrintfLine("500 error in LISTGROUP: %s", err.Error())
						}
					} else {
						// don't has newsgroup
						conn.PrintfLine("411 no such newsgroup")
					}
				} else {
					conn.PrintfLine("412 no newsgroup selected")
				}
			} else if cmd == "NEWSGROUPS" {
				// handle NEWSGROUPS
				conn.PrintfLine("231 List of newsgroups follow")
				dw := conn.DotWriter()
				// get a list of every newsgroup
				groups := daemon.database.GetAllNewsgroups()
				// for each group
				for _, group := range groups {
					// get low/high water mark
					// XXX: heavy as shit
					//lo, hi, err := daemon.database.GetLastAndFirstForGroup(group)
					//if err == nil {
					if ValidNewsgroup(group) {
						fmt.Fprintf(dw, "%s 0 0 y\n", group)
					}
					//} else {
					//	log.Println(self.name, "could not get low/high water mark for", group, err)
					//}
				}
				// flush dotwriter
				dw.Close()

			} else if cmd == "XOVER" {
				// handle XOVER
				if self.group == "" {
					conn.PrintfLine("412 No newsgroup selected")
				} else {
					// handle xover command
					// right now it's every article in group
					models, err := daemon.database.GetNNTPPostsInGroup(self.group)
					if err == nil {
						conn.PrintfLine("224 Overview information follows")
						dw := conn.DotWriter()
						for _, model := range models {
							if model != nil {
								if err == nil {
									/*
										The first 8 fields MUST be the following, in order:
										     "0" or article number (see below)
										     Subject header content
										     From header content
										     Date header content
										     Message-ID header content
										     References header content
										     :bytes metadata item
										     :lines metadata item
									*/
									fmt.Fprintf(dw,
										"%.6d\t%s\t\"%s\" <%s@%s>\t%s\t%s\t%s\r\n",
										model.NNTPID(),
										safeHeader(model.Subject()),
										safeHeader(model.Name()), safeHeader(model.Name()), safeHeader(model.Frontend()),
										safeHeader(model.Date()),
										safeHeader(model.MessageID()),
										safeHeader(model.Reference()))
								}
							}
						}
						dw.Close()
					} else {
						log.Println(self.name, "error when getting posts in", self.group, err)
						conn.PrintfLine("500 error, %s", err.Error())
					}
				}
			} else if cmd == "HEAD" {
				if len(self.group) == 0 {
					// no group selected
					conn.PrintfLine("412 No newsgroup slected")
				} else {
					// newsgroup is selected
					// handle HEAD command
					if len(parts) == 0 {
						// we have no parameters
						if len(self.selected_article) > 0 {
							// we have a selected article
						} else {
							// no selected article
							conn.PrintfLine("420 current article number is invalid")
						}
					} else {
						// head command has 1 or more paramters
						var n int64
						var msgid string
						var has bool
						var code int
						n, err = strconv.ParseInt(parts[1], 10, 64)
						if err == nil {
							// is a number
							msgid, err = daemon.database.GetMessageIDForNNTPID(self.group, n)
							if err == nil && len(msgid) > 0 {
								has = daemon.store.HasArticle(msgid)
							}
							if !has {
								code = 423
							}
						} else if ValidMessageID(parts[1]) {
							msgid = parts[1]
							has = daemon.store.HasArticle(msgid)
							if has {
								n, err = daemon.database.GetNNTPIDForMessageID(self.group, parts[1])
							} else {
								code = 430
							}
						}
						if err == nil {
							if has {
								// we has
								hdrs := daemon.store.GetHeaders(msgid)
								if hdrs == nil {
									// wtf can't load?
									conn.PrintfLine("500 cannot load headers")
								} else {
									// headers loaded, send them
									conn.PrintfLine("221 %d %s", n, msgid)
									dw := conn.DotWriter()
									err = writeMIMEHeader(dw, hdrs)
									dw.Close()
									hdrs = nil
								}
							} else if code > 0 {
								// don't has
								conn.PrintfLine("%d don't have article", code)
							} else {
								// invalid state
								conn.PrintfLine("500 invalid state in HEAD, should have article but we don't")
							}
						} else {
							// error occured
							conn.PrintfLine("500 error in HEAD: %s", err.Error())
						}
					}
				}
			} else if cmd == "GROUP" {
				// handle GROUP command
				group := parts[1]
				// check for newsgroup
				if daemon.database.HasNewsgroup(group) {
					// we have the group
					self.group = group
					// count posts
					number := daemon.database.CountPostsInGroup(group, 0)
					// get hi/low water marks
					hi, low, err := daemon.database.GetLastAndFirstForGroup(group)
					if err == nil {
						// we gud
						conn.PrintfLine("211 %d %d %d %s", number, low, hi, group)
					} else {
						// wtf error
						log.Println(self.name, "error in GROUP command", err)
						// still have to reply, send it bogus low/hi
						conn.PrintfLine("211 %d 0 1 %s", number, group)
					}
				} else {
					// no such group
					conn.PrintfLine("411 No Such Newsgroup")
				}
			} else if cmd == "LIST" && parts[1] == "NEWSGROUPS" {
				// handle list command
				list, err := daemon.database.GetNewsgroupList()
				if err == nil {
					conn.PrintfLine("215 list of newsgroups follows")
					dw := conn.DotWriter()
					for _, entry := range list {
						if ValidNewsgroup(entry[0]) {
							io.WriteString(dw, fmt.Sprintf("%s %s %s y\r\n", entry[0], entry[1], entry[2]))
						}
					}
					dw.Close()
				} else {
					conn.PrintfLine("500 failed to get list: %s", err.Error())
				}
			} else if cmd == "STAT" {
				if len(self.group) == 0 {
					if len(parts) == 2 {
						// parameter given
						msgid := parts[1]
						// check for article
						if ValidMessageID(msgid) && daemon.store.HasArticle(msgid) {
							// valid message id
							var n int64
							n, err = daemon.database.GetNNTPIDForMessageID(self.group, msgid)
							// exists
							conn.PrintfLine("223 %d %s", n, msgid)
							err = nil
						} else {
							conn.PrintfLine("430 No article with that message-id")
						}
					} else {
						conn.PrintfLine("412 No newsgroup selected")
					}
				} else if daemon.database.HasNewsgroup(self.group) {
					// group specified
					if len(parts) == 2 {
						// parameter specified
						var msgid string
						var n int64
						n, err = strconv.ParseInt(parts[1], 10, 64)
						if err == nil {
							msgid, err = daemon.database.GetMessageIDForNNTPID(self.group, n)
							if err != nil {
								// error getting id
								conn.PrintfLine("500 error getting nntp article id: %s", err.Error())
								return
							}
						} else {
							// message id
							msgid = parts[1]
						}
						if ValidMessageID(msgid) && daemon.store.HasArticle(msgid) {
							conn.PrintfLine("223 %d %s", n, msgid)
						} else if n == 0 {
							// was a message id
							conn.PrintfLine("430 no such article")
						} else {
							// was an article number
							conn.PrintfLine("423 no article with that number")
						}
					} else {
						conn.PrintfLine("420 Current article number is invalid")
					}
				} else {
					conn.PrintfLine("500 invalid daemon state, got STAT with group set but we don't have that group now?")
				}
			} else if cmd == "XPAT" {
				var hdr string
				var msgid string
				var lo, hi int64
				var pats []string
				if len(parts) >= 3 {
					hdr = parts[1]
					if ValidMessageID(parts[2]) {
						msgid = parts[2]
					} else {
						lo, hi = parseRange(parts[2])
						if !ValidNewsgroup(self.group) {
							conn.PrintfLine("430 no such article")
							return
						}
					}
					pats = parts[3:]
					var hdrs ArticleHeaders

					if len(msgid) > 0 {
						hdrs, err = daemon.database.GetHeadersForMessage(msgid)
					} else {
						hdrs, err = daemon.database.FindHeaders(self.group, hdr, lo, hi)
					}
					if err == nil {
						hdrs = headerFindPats(hdr, hdrs, pats)
						if hdrs.Len() > 0 {
							conn.PrintfLine("221 Header follows")
							for _, vals := range hdrs {
								for idx := range vals {
									conn.PrintfLine("%s: %s", hdr, vals[idx])
								}
							}
							conn.PrintfLine(".")
						} else {
							conn.PrintfLine("430 no such article")
						}
					} else {
						conn.PrintfLine("502 %s", err.Error())
					}
					return
				}
				conn.PrintfLine("430 no such article")
				return
			} else if cmd == "XHDR" {
				if len(self.group) > 0 {
					var msgid string
					var hdr string
					if len(parts) == 2 {
						// XHDR headername
						hdr = parts[1]
						if self.selected_article == "" {
							conn.PrintfLine("420 No Current Article selected")
							return
						}
						msgid = self.selected_article
					} else if len(parts) == 3 {
						// message id
						msgid = parts[2]
						hdr = parts[1]
					} else {
						// wtf?
						conn.PrintfLine("502 no permission")
						return
					}
					if ValidMessageID(msgid) {
						hdrs := daemon.store.GetHeaders(msgid)
						if hdrs != nil {
							v := hdrs.Get(hdr, "")
							conn.PrintfLine("221 header follows")
							conn.PrintfLine(v)
							conn.PrintfLine(".")
						} else {
							conn.PrintfLine("500 could not fetch headers for %s", msgid)
						}
					} else {
						conn.PrintfLine("430 no such article")
					}
				} else {
					// no newsgroup
					conn.PrintfLine("412 no newsgroup selected")
				}

			} else {
				log.Println(self.name, "invalid command recv'd", cmd)
				conn.PrintfLine("500 Invalid command: %s", cmd)
			}
		} else {
			if line == "LIST" {
				list, err := daemon.database.GetNewsgroupList()
				if err == nil {
					conn.PrintfLine("215 list of newsgroups follows")
					dw := conn.DotWriter()
					for _, entry := range list {
						if ValidNewsgroup(entry[0]) {
							io.WriteString(dw, fmt.Sprintf("%s %s %s y\r\n", entry[0], entry[2], entry[1]))
						}
					}
					dw.Close()
				} else {
					conn.PrintfLine("500 failed to get list: %s", err.Error())
				}
			} else if line == "POST" {
				if !self.authenticated {
					// needs tls to work if not logged in
					conn.PrintfLine("440 Posting Not Allowed")
				} else {
					// handle POST command
					conn.PrintfLine("340 Yeeeh postit yo; end with <CR-LF>.<CR-LF>")
					var msg *mail.Message
					msg, err = readMIMEHeader(bufio.NewReader(conn.DotReader()))
					var success bool
					var reason string
					if err == nil {
						hdr := textproto.MIMEHeader(msg.Header)
						if getMessageID(hdr) == "" {
							hdr.Set("Message-ID", genMessageID(daemon.instance_name))
						}
						msgid = getMessageID(hdr)
						hdr.Set("Date", timeNowStr())
						ipaddr, _, _ := net.SplitHostPort(self.addr.String())
						if len(ipaddr) > 0 {
							// inject encrypted ip for poster
							encaddr, err := daemon.database.GetEncAddress(ipaddr)
							if err == nil {
								hdr.Set("X-Encrypted-Ip", encaddr)
							}
						}
						reason, _, err = self.checkMIMEHeader(daemon, hdr)
						success = reason == "" && err == nil
						if success {
							refs := strings.Split(hdr.Get("References"), " ")
							newsgroup := hdr.Get("Newsgroups")
							for _, reference := range refs {
								if reference != "" && ValidMessageID(reference) {
									if !daemon.store.HasArticle(reference) && !daemon.database.IsExpired(reference) {
										log.Println(self.name, "got reply to", reference, "but we don't have it")
										go daemon.askForArticle(reference)
									} else {
										h := daemon.store.GetMIMEHeader(reference)
										if strings.Trim(h.Get("References"), " ") == "" {
											hdr.Set("References", getMessageID(h))
										}
									}
								} else if reference != "" {
									// bad message id
									reason = "cannot reply with invalid reference, maybe you are replying to a reply?"
									success = false
								}
							}
							if success && daemon.database.HasNewsgroup(newsgroup) {
								body := &io.LimitedReader{
									R: msg.Body,
									N: daemon.messageSizeLimitFor(newsgroup),
								}
								err = storeMessage(daemon, hdr, body)
							}
						}
					}
					if success {
						// all gud
						conn.PrintfLine("240 We got it, thnkxbai")
					} else {
						// failed posting
						if err != nil {
							log.Println(self.name, "failed nntp POST", err)
							reason = err.Error()
						}
						conn.PrintfLine("441 Posting Failed %s", reason)
					}
				}
			} else {
				conn.PrintfLine("500 wut?")
			}
		}
	}
	return
}

func (self *nntpConnection) startStreaming(daemon *NNTPDaemon, reader bool, conn *textproto.Conn) {
	self.keepalive = time.NewTicker(time.Minute)
	defer self.keepalive.Stop()
	err := self.handleStreaming(daemon, conn)
	if err == nil {
		log.Println(self.name, "done with streaming")
	} else {
		log.Println(self.name, "error while streaming:", err)
	}
}

// interprets result string from response code 211 of GROUP message
// for parts it fails, returns zeros
func interpretGroupResult(line string) (es, lo, hi uint64, group string) {
	// this code is braindead but I was lazy to search stdlib
	startes := 0
	for startes < len(line) && (line[startes] == ' ' || line[startes] == '\t') {
		startes++
	}
	endes := startes
	for endes < len(line) && line[endes] != ' ' && line[endes] != '\t' {
		endes++
	}
	startlo := endes
	for startlo < len(line) && (line[startlo] == ' ' || line[startlo] == '\t') {
		startlo++
	}
	endlo := startlo
	for endlo < len(line) && line[endlo] != ' ' && line[endlo] != '\t' {
		endlo++
	}
	starthi := endlo
	for starthi < len(line) && (line[starthi] == ' ' || line[starthi] == '\t') {
		starthi++
	}
	endhi := starthi
	for endhi < len(line) && line[endhi] != ' ' && line[endhi] != '\t' {
		endhi++
	}
	startgroup := endhi
	for startgroup < len(line) && (line[startgroup] == ' ' || line[startgroup] == '\t') {
		startgroup++
	}
	// will return 0 if failed to parse. which is OK for us
	es, _ = strconv.ParseUint(line[startes:endes], 10, 64)
	lo, _ = strconv.ParseUint(line[startlo:endlo], 10, 64)
	hi, _ = strconv.ParseUint(line[starthi:endhi], 10, 64)
	group = line[startgroup:]
	return
}

const maxXOVERRange = 800

// scrape all posts in a newsgroup
// download ones we do not have
func (self *nntpConnection) scrapeGroup(daemon *NNTPDaemon, conn *textproto.Conn, group string) (err error) {

	log.Println(self.name, "scrape newsgroup", group)
	// send GROUP command
	err = conn.PrintfLine("GROUP %s", group)
	if err == nil {
		// read reply to GROUP command
		var code int
		var ret string
		code, ret, err = conn.ReadCodeLine(211)
		// check code
		if code == 211 {
			// success
			es, lo, hi, _ := interpretGroupResult(ret)
			for {
				if lo > hi {
					// server indicated empty group
					// or
					// we finished pulling stuff
					break
				}
				if hi-lo+1 <= maxXOVERRange {
					// not too much for us to pull
					if es == 0 && lo == 0 && hi == 0 {
						// empty group
						break
					}
					if lo != hi {
						// usual normal case
						err = conn.PrintfLine("XOVER %d-%d", lo, hi)
					} else if lo == 0 {
						// probably something went wrong. try pulling more
						err = conn.PrintfLine("XOVER 0-")
					} else {
						// normal case with one article
						err = conn.PrintfLine("XOVER %d", lo)
					}
					lo = hi + 1
				} else {
					// too much to pull in one shot
					err = conn.PrintfLine("XOVER %d-%d", lo, lo+maxXOVERRange-1)
					lo += maxXOVERRange
				}
				if err != nil {
					// something went very wrong
					return
				}
				// no error sending command, read first line
				code, _, err = conn.ReadCodeLine(224)
				if code == 224 {
					// maps message-id -> references
					articles := make(map[string]string)
					// successful response, read multiline
					dr := conn.DotReader()
					sc := bufio.NewScanner(dr)
					for sc.Scan() {
						line := sc.Text()
						parts := strings.Split(line, "\t")
						if len(parts) > 5 {
							// probably valid line
							msgid := parts[4]
							// msgid -> reference
							articles[msgid] = parts[5]
							// incase server returned more articles than we requested
							if num, nerr := strconv.ParseUint(parts[0], 10, 64); nerr == nil && num >= lo {
								// fix lo so that we wont request them again
								lo = num + 1
							}
						} else {
							// probably not valid line
							// ignore
						}
					}
					err = sc.Err()
					if err == nil {
						// everything went okay when reading multiline
						// for each article
						for msgid, refid := range articles {
							// check the reference
							if len(refid) > 0 && ValidMessageID(refid) {
								// do we have it?
								if daemon.database.HasArticle(refid) {
									// we have it don't do anything
								} else if daemon.database.ArticleBanned(refid) {
									// thread banned
								} else {
									// we don't got root post and it's not banned, try getting it
									err = self.requestArticle(daemon, conn, refid)
									if err != nil {
										// something bad happened
										log.Println(self.name, "failed to obtain root post", refid, err)
										// it fails only when REALLY bad stuff happens
										return
									}
								}
							}
							// check the actual message-id
							if len(msgid) > 0 && ValidMessageID(msgid) {
								// do we have it?
								if daemon.database.HasArticle(msgid) {
									// we have it, don't do shit
								} else if daemon.database.ArticleBanned(msgid) {
									// this article is banned, don't do shit yo
								} else {
									// we don't have it but we want it
									err = self.requestArticle(daemon, conn, msgid)
									if err != nil {
										// something bad happened
										log.Println(self.name, "failed to obtain article", msgid, err)
										// it fails only when REALLY bad stuff happens
										return
									}
								}
							}
						}
					} else {
						// something bad went down when reading multiline
						log.Println(self.name, "failed to read multiline for", group, "XOVER command")
					}
				} else if code >= 420 || code <= 429 {
					// group doesn't have articles in one way or another. not really error
					err = nil
				}
			}
		} else if err == nil {
			// invalid response code no error
			log.Println(self.name, "says they don't have", group, "but they should")
		} else {
			// error recving response
			log.Println(self.name, "error recving response from GROUP command", err)
		}
	}
	return
}

// ask for an article from the remote server
func (self *nntpConnection) askForArticle(msgid string) {
	if self.messageIsQueued(msgid) {
		// already queued
	} else {
		self.articles_access.Lock()
		log.Println(self.name, "asking for", msgid)
		self.messageSetPendingState(msgid, "queued", 0)
		self.articles = append(self.articles, msgid)
		self.articles_access.Unlock()
	}
}

// grab every post from the remote server, assumes outbound connection
func (self *nntpConnection) scrapeServer(daemon *NNTPDaemon, conn *textproto.Conn) (err error) {
	self.abort = func() {
		conn.Close()
	}

	defer func() {
		self.abort = nil
	}()
	log.Println(self.name, "scrape remote server")
	success := true
	if success {
		// send newsgroups command
		err = conn.PrintfLine("LIST NEWSGROUPS")
		if err == nil {
			// read response line
			code, line, err := conn.ReadCodeLine(0)
			if code == 231 || code == 215 {
				var groups []string
				// valid response, we expect a multiline
				dr := conn.DotReader()
				// read lines
				sc := bufio.NewScanner(dr)
				for sc.Scan() {
					line := sc.Text()
					idx := strings.IndexAny(line, " \t")
					if idx > 0 {
						log.Println(self.name, "got newsgroup", line[:idx])
						groups = append(groups, line[:idx])
					} else if idx < 0 {
						log.Println(self.name, "got newsgroup", line)
						groups = append(groups, line)
					} else {
						// can't have it starting with WS
						log.Printf("%s invalid line in newsgroups multiline response [%s]\n", self.name, line)
					}
				}
				err = sc.Err()
				if err == nil {
					log.Println(self.name, "got list of newsgroups")
					// for each group
					for _, group := range groups {
						var banned bool
						// check if the newsgroup is banned
						banned, err = daemon.database.NewsgroupBanned(group)
						if banned {
							// we don't want it
						} else if err == nil {
							if daemon.AllowsNewsgroup(group) {
								// scrape the group
								err = self.scrapeGroup(daemon, conn, group)
								if err != nil {
									log.Println(self.name, "failure scraping group", group, "error:", err)
									// do not break here, continue with other groups
								}
							}
						} else {
							// error while checking for ban
							log.Println(self.name, "checking for newsgroup ban failed", err)
							break
						}
					}
				} else {
					// we got a bad multiline block?
					log.Println(self.name, "bad multiline response from newsgroups command", err)
				}
			} else if err == nil {
				// invalid response no error
				log.Println(self.name, "gave us invalid response to newsgroups command", line)
			} else {
				// invalid response with error
				log.Println(self.name, "error while reading response from newsgroups command", err)
			}
		} else {
			log.Println(self.name, "failed to send newsgroups command", err)
		}
	} else if err == nil {
		// failed to switch mode to reader
		log.Println(self.name, "does not do reader mode, bailing scrape")
	} else {
		// failt to switch mode because of error
		log.Println(self.name, "failed to switch to reader mode when scraping", err)
	}
	return
}

// ask for an article from the remote server
// feed it to the daemon if we get it
func (self *nntpConnection) requestArticle(daemon *NNTPDaemon, conn *textproto.Conn, msgid string) (err error) {
	if daemon.store.HasArticle(msgid) {
		return
	}
	// send command
	err = conn.PrintfLine("ARTICLE %s", msgid)
	// read response
	var code int
	var line string
	code, line, err = conn.ReadCodeLine(-1)
	if code == 220 {
		// awwww yeh we got it
		var msg *mail.Message
		// read header
		msg, err = readMIMEHeader(bufio.NewReader(conn.DotReader()))
		if err == nil {
			hdr := textproto.MIMEHeader(msg.Header)
			// prepare to read body
			// check header and decide if we want this
			reason, ban, err := self.checkMIMEHeaderNoAuth(daemon, hdr)
			if err == nil {
				if len(reason) > 0 {
					log.Println(self.name, "discarding", msgid, reason)
					// we don't want it, discard
					io.Copy(ioutil.Discard, msg.Body)
					if ban {
						daemon.database.BanArticle(msgid, reason)
					}
				} else {
					// yeh we want it open up a file to store it in
					body := &io.LimitedReader{
						R: msg.Body,
						N: daemon.messageSizeLimitFor(hdr.Get("Newsgroups")),
					}
					err = storeMessage(daemon, hdr, body)
					if err != nil {
						log.Println(self.name, "failed to obtain article", err)
						daemon.database.BanArticle(msgid, err.Error())
					}
				}
			} else {
				// error happened while processing
				log.Println(self.name, "error happend while processing MIME header", err)
			}
		} else {
			// error happened while reading header
			log.Println(self.name, "error happened while reading MIME header", err)
		}
	} else if code == 430 {
		// they don't know it D:
	} else {
		// invalid response
		log.Println(self.name, "invald response to ARTICLE:", code, line)
	}
	if err != nil {
		log.Println(self.name, err.Error())
		conn.Close()
	}
	return
}

func (self *nntpConnection) startReader(daemon *NNTPDaemon, conn *textproto.Conn) {
	log.Println(self.name, "run reader mode")
	var err error
	for err == nil {
		select {
		case chnl := <-self.die:
			// we were asked to die
			// send quit
			conn.PrintfLine("QUIT")
			chnl <- true
			break
		default:
			var msgid string
			self.articles_access.Lock()
			if len(self.articles) > 0 {
				msgid = self.articles[0]
				self.articles = self.articles[1:]
			}
			self.articles_access.Unlock()
			if len(msgid) > 0 {
				// next article to ask for
				log.Println(self.name, "obtaining", msgid)
				self.messageSetPendingState(msgid, "article", 0)
				err = self.requestArticle(daemon, conn, msgid)
				self.messageSetProcessed(msgid)
				if err != nil {
					log.Println(self.name, "error while in reader mode:", err)
				}
			} else {
				time.Sleep(time.Second)
			}
		}
	}
	// close connection
	conn.Close()
}

// run the mainloop for this connection
// stream if true means they support streaming mode
// reader if true means they support reader mode
func (self *nntpConnection) runConnection(daemon *NNTPDaemon, inbound, stream, reader, use_tls bool, preferMode string, nconn net.Conn, conf *FeedConfig) {
	defer nconn.Close()
	self.addr = nconn.RemoteAddr()
	var err error
	var line string
	var success bool
	var conn *textproto.Conn

	if (conf != nil && !conf.tls_off) && use_tls && daemon.CanTLS() && !inbound {
		log.Println(self.name, "STARTTLS with", self.hostname)
		conn, _, err = SendStartTLS(nconn, daemon.GetTLSConfig(self.hostname))
		if err != nil {
			log.Println(self.name, err)
			return
		}

	} else {
		// we are authenticated if we are don't need tls
		conn = textproto.NewConn(nconn)
	}
	if !inbound {
		if preferMode == "stream" {
			// try outbound streaming
			if stream {
				success, err = self.modeSwitch("STREAM", conn)
				if success {
					self.mode = "STREAM"
					// start outbound streaming in background
					go self.startStreaming(daemon, reader, conn)
				}
			}
		} else if reader {
			// try reader mode
			success, err = self.modeSwitch("READER", conn)
			if success {
				self.mode = "READER"
				self.startReader(daemon, conn)
				return
			}
		}
		if success {
			log.Println(self.name, "mode set to", self.mode)
		} else {
			// bullshit
			// we can't do anything so we quit
			log.Println(self.name, "can't stream or read, wtf?", err)
			conn.PrintfLine("QUIT")
			conn.Close()
			return
		}
	}

	for err == nil {
		line, err = conn.ReadLine()
		if inbound && strings.HasPrefix(line, "QUIT") {
			conn.PrintfLine("205 bai")
			conn.Close()
			return
		}
		if self.mode == "" {
			if inbound {
				if len(line) == 0 {
					conn.Close()
					return
				}
				parts := strings.Split(line, " ")
				cmd := parts[0]
				if cmd == "STARTTLS" {
					_conn, state, err := HandleStartTLS(nconn, daemon.GetOurTLSConfig())
					if err == nil {
						// we are now tls
						conn = _conn
						self.tls_state = state
						self.authenticated = true
						log.Println(self.name, "TLS initiated", self.authenticated)
					} else {
						log.Println("STARTTLS failed:", err)
						nconn.Close()
						return
					}
				} else if cmd == "CAPABILITIES" {
					// write capabilities
					conn.PrintfLine("101 i support to the following:")
					dw := conn.DotWriter()
					caps := []string{"VERSION 2", "READER", "STREAMING", "IMPLEMENTATION srndv2", "POST", "IHAVE", "AUTHINFO"}
					if daemon.CanTLS() {
						caps = append(caps, "STARTTLS")
					}
					for _, cap := range caps {
						io.WriteString(dw, cap)
						io.WriteString(dw, "\n")
					}
					dw.Close()
					log.Println(self.name, "sent Capabilities")
				} else if cmd == "MODE" {
					if len(parts) == 2 {
						mode := strings.ToUpper(parts[1])
						if mode == "READER" {
							// set reader mode
							self.mode = "READER"
							// we'll allow posting for reader
							conn.PrintfLine("200 Posting is Permitted awee yeh")
						} else if mode == "STREAM" {
							if !self.authenticated {
								conn.PrintfLine("483 Streaming Denied")
							} else {
								// set streaming mode
								conn.PrintfLine("203 Stream it brah")
								self.mode = "STREAM"
								log.Println(self.name, "streaming enabled")
							}
						}
					}
				} else {
					// handle a it as a command, we don't have a mode set
					parts := strings.Split(line, " ")
					cmd := parts[0]
					if cmd == "STARTTLS" {
						_conn, state, err := HandleStartTLS(nconn, daemon.GetOurTLSConfig())
						if err == nil {
							// we are now tls
							conn = _conn
							self.tls_state = state
							self.authenticated = state.HandshakeComplete
							log.Println("TLS initiated")
						} else {
							log.Println("STARTTLS failed:", err)
							nconn.Close()
							return
						}
					}
					var code64 int64
					code64, err = strconv.ParseInt(parts[0], 10, 32)
					if err == nil {
						err = self.handleLine(daemon, int(code64), line[4:], conn)
					} else {
						err = self.handleLine(daemon, 0, line, conn)
					}
				}
			}
		} else {
			if err == nil {
				parts := strings.Split(line, " ")
				var code64 int64
				code64, err = strconv.ParseInt(parts[0], 10, 32)
				if err == nil {
					err = self.handleLine(daemon, int(code64), line[4:], conn)
				} else {
					err = self.handleLine(daemon, 0, line, conn)
				}
			}
		}
	}
	if err != io.EOF {
		log.Println(self.name, "got error", err)
		if !inbound && conn != nil {
			// send quit on outbound
			conn.PrintfLine("QUIT")
		}
	}
	nconn.Close()
}
