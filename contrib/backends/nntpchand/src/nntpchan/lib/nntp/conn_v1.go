package nntp

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"io"
	"mime"
	"mime/multipart"
	"net"
	"net/textproto"
	"nntpchan/lib/config"
	"nntpchan/lib/nntp/message"
	"nntpchan/lib/store"
	"nntpchan/lib/util"
	"os"
	"strings"
)

// handles 1 line of input from a connection
type lineHandlerFunc func(c *v1Conn, line string, hooks EventHooks) error

// base nntp connection
type v1Conn struct {
	// buffered connection
	C *textproto.Conn

	// unexported fields ...

	// connection state (mutable)
	state ConnState
	// tls connection if tls is established
	tlsConn *tls.Conn
	// tls config for this connection, nil if we don't support tls
	tlsConfig *tls.Config
	// has this connection authenticated yet?
	authenticated bool
	// the username logged in with if it has authenticated via user/pass
	username string
	// underlying network socket
	conn net.Conn
	// server's name
	serverName string
	// article acceptor checks if we want articles
	acceptor ArticleAcceptor
	// headerIO for read/write of article header
	hdrio *message.HeaderIO
	// article storage
	storage store.Storage
	// event callbacks
	hooks EventHooks
	// inbound connection authenticator
	auth ServerAuth
	// command handlers
	cmds map[string]lineHandlerFunc
}

// json representation of this connection
// format is:
// {
//   "state" : (connection state object),
//   "authed" : bool,
//   "tls" : (tls info or null if plaintext connection)
// }
func (c *v1Conn) MarshalJSON() ([]byte, error) {
	j := make(map[string]interface{})
	j["state"] = c.state
	j["authed"] = c.authenticated
	if c.tlsConn == nil {
		j["tls"] = nil
	} else {
		j["tls"] = c.tlsConn.ConnectionState()
	}
	return json.Marshal(j)
}

// get the current state of our connection (immutable)
func (c *v1Conn) GetState() (state *ConnState) {
	return &ConnState{
		FeedName: c.state.FeedName,
		ConnName: c.state.ConnName,
		HostName: c.state.HostName,
		Mode:     c.state.Mode,
		Group:    c.state.Group,
		Article:  c.state.Article,
		Policy: &FeedPolicy{
			Whitelist:            c.state.Policy.Whitelist,
			Blacklist:            c.state.Policy.Blacklist,
			AllowAnonPosts:       c.state.Policy.AllowAnonPosts,
			AllowAnonAttachments: c.state.Policy.AllowAnonAttachments,
			AllowAttachments:     c.state.Policy.AllowAttachments,
			UntrustedRequiresPoW: c.state.Policy.UntrustedRequiresPoW,
		},
	}
}

func (c *v1Conn) Group() string {
	return c.state.Group.String()
}

func (c *v1Conn) IsOpen() bool {
	return c.state.Open
}

func (c *v1Conn) Mode() Mode {
	return c.state.Mode
}

// is posting allowed rignt now?
func (c *v1Conn) PostingAllowed() bool {
	return c.Authed()
}

// process incoming commands
// call event hooks as needed
func (c *v1Conn) Process(hooks EventHooks) {
	var err error
	var line string
	for err == nil {
		line, err = c.readline()
		if len(line) == 0 {
			// eof (proably?)
			c.Close()
			return
		}

		uline := strings.ToUpper(line)
		parts := strings.Split(uline, " ")
		handler, ok := c.cmds[parts[0]]
		if ok {
			// we know the command
			err = handler(c, line, hooks)
		} else {
			// we don't know the command
			err = c.printfLine("%s Unknown Command: %s", RPL_UnknownCommand, line)
		}
	}
}

type v1OBConn struct {
	C               v1Conn
	supports_stream bool
	streamChnl      chan ArticleEntry
	conf            *config.FeedConfig
}

func (c *v1OBConn) IsOpen() bool {
	return c.IsOpen()
}

func (c *v1OBConn) Mode() Mode {
	return c.Mode()
}

func (c *v1OBConn) DownloadGroup(g Newsgroup) (err error) {
	err = c.C.printfLine(CMD_Group(g).String())
	if err == nil {
		var line string
		line, err = c.C.readline()
		if strings.HasPrefix(line, RPL_NoSuchGroup) {
			// group does not exist
			// don't error this is not a network io error
			return
		}
		// send XOVER
		err = c.C.printfLine(CMD_XOver.String())
		if err == nil {
			line, err = c.C.readline()
			if err == nil {
				if !strings.HasPrefix(line, RPL_Overview) {
					// bad response
					// not a network io error, don't error
					return
				}
				var msgids []MessageID
				// read reply
				for err == nil && line != "." {
					line, err = c.C.readline()
					parts := strings.Split(line, "\t")
					if len(parts) != 6 {
						// incorrect size
						continue
					}
					m := MessageID(parts[4])
					r := MessageID(parts[5])
					if c.C.acceptor == nil {
						// no acceptor take it if store doesn't have it
						if c.C.storage.HasArticle(m.String()) == store.ErrNoSuchArticle {
							msgids = append(msgids, m)
						}
					} else {
						// check if thread is banned
						if c.C.acceptor.CheckMessageID(r).Ban() {
							continue
						}
						// check if message is wanted
						if c.C.acceptor.CheckMessageID(m).Accept() {
							msgids = append(msgids, m)
						}
					}
				}
				var accepted []MessageID

				for _, msgid := range msgids {

					if err != nil {
						return // io error
					}

					if !msgid.Valid() {
						// invalid message id
						continue
					}
					// get message header
					err = c.C.printfLine(CMD_Head(msgid).String())
					if err == nil {
						line, err = c.C.readline()
						if err == nil {
							if !strings.HasPrefix(line, RPL_ArticleHeaders) {
								// bad response
								continue
							}
							// read message header
							dr := c.C.C.DotReader()
							var hdr message.Header
							hdr, err = c.C.hdrio.ReadHeader(dr)
							if err == nil {
								if c.C.acceptor == nil {
									accepted = append(accepted, msgid)
								} else if c.C.acceptor.CheckHeader(hdr).Accept() {
									accepted = append(accepted, msgid)
								}
							}
						}
					}
				}
				// download wanted messages
				for _, msgid := range accepted {
					if err != nil {
						// io error
						return
					}
					// request message
					err = c.C.printfLine(CMD_Article(msgid).String())
					if err == nil {
						line, err = c.C.readline()
						if err == nil {
							if !strings.HasPrefix(line, RPL_Article) {
								// bad response
								continue
							}
							// read article
							_, err = c.C.readArticle(false, c.C.hooks)
							if err == nil {
								// we read it okay
							}
						}
					}
				}
			}
		}
	}
	return
}

func (c *v1OBConn) ListNewsgroups() (groups []Newsgroup, err error) {
	err = c.C.printfLine(CMD_Newsgroups.String())
	if err == nil {
		var line string
		line, err = c.C.readline()
		if err == nil {
			if !strings.HasPrefix(line, RPL_NewsgroupList) {
				// bad stuff
				err = errors.New("invalid reply for NEWSGROUPS command: " + line)
				return
			}
			for err == nil && line != "." {
				line, err = c.C.readline()
				if err == nil {
					parts := strings.Split(line, " ")
					if len(parts) != 4 {
						// bad format
						continue
					}
					groups = append(groups, Newsgroup(parts[0]))
				}
			}
		}
	}
	return
}

// negioate outbound connection
func (c *v1OBConn) Negotiate(stream bool) (err error) {
	var line string
	// discard first line
	_, err = c.C.readline()
	if err == nil {
		// request capabilities
		err = c.C.printfLine(CMD_Capabilities.String())
		dr := c.C.C.DotReader()
		var b bytes.Buffer
		_, err = io.Copy(&b, dr)
		if err == nil {
			// try login if specified
			if c.conf.Username != "" && c.conf.Password != "" {
				err = c.C.printfLine("AUTHINFO USER %s", c.conf.Username)
				if err != nil {
					return
				}
				line, err = c.C.readline()
				if strings.HasPrefix(line, RPL_MoreAuth) {
					err = c.C.printfLine("AUTHINFO PASS %s", c.conf.Password)
					if err != nil {
						return
					}
					line, err = c.C.readline()
					if err != nil {
						return
					}
					if strings.HasPrefix(line, RPL_AuthAccepted) {
						log.WithFields(log.Fields{
							"name": c.conf.Name,
							"user": c.conf.Username,
						}).Info("authentication accepted")
					} else {
						// not accepted?
						err = errors.New(line)
					}
				} else {
					// bad user?
					err = errors.New(line)
				}
			}
			if err == nil {
				if stream {
					// set mode stream
					err = c.C.printfLine(ModeStream.String())
					if err == nil {
						line, err = c.C.readline()
						if err == nil && !strings.HasPrefix(line, RPL_PostingStreaming) {
							err = errors.New("streaiming not allowed")
						}
					}
				}
			}
		}
	}
	return
}

func (c *v1OBConn) PostingAllowed() bool {
	return c.C.PostingAllowed()
}

func (c *v1OBConn) ProcessInbound(hooks EventHooks) {

}

func (c *v1OBConn) WantsStreaming() bool {
	return c.supports_stream
}

func (c *v1OBConn) StreamAndQuit() {
	for {
		e, ok := <-c.streamChnl
		if ok {
			// do CHECK
			msgid := e.MessageID()
			if !msgid.Valid() {
				log.WithFields(log.Fields{
					"pkg":   "nntp-conn",
					"state": c.C.state,
					"msgid": msgid,
				}).Warn("Dropping stream event with invalid message-id")
				continue
			}
			// send line
			err := c.C.printfLine("%s %s", stream_CHECK, msgid)
			if err == nil {
				// read response
				var line string
				line, err = c.C.readline()
				ev := StreamEvent(line)
				if ev.Valid() {
					cmd := ev.Command()
					if cmd == RPL_StreamingAccept {
						// accepted to send

						// check if we really have it in storage
						err = c.C.storage.HasArticle(msgid.String())
						if err == nil {
							var r io.ReadCloser
							r, err = c.C.storage.OpenArticle(msgid.String())
							if err == nil {
								log.WithFields(log.Fields{
									"pkg":   "nntp-conn",
									"state": c.C.state,
									"msgid": msgid,
								}).Debug("article accepted will send via TAKETHIS now")
								_ = c.C.printfLine("%s %s", stream_TAKETHIS, msgid)
								br := bufio.NewReader(r)
								n := int64(0)
								for err == nil {
									var line string
									line, err = br.ReadString(10)
									if err == io.EOF {
										err = nil
										break
									}
									line = strings.Trim(line, "\r")
									line = strings.Trim(line, "\n")
									err = c.C.printfLine(line)
									n += int64(len(line))
								}
								r.Close()
								err = c.C.printfLine(".")
								if err == nil {
									// successful takethis sent
									log.WithFields(log.Fields{
										"pkg":   "nntp-conn",
										"state": c.C.state,
										"msgid": msgid,
										"bytes": n,
									}).Debug("article transfer done")
									// read response
									line, err = c.C.readline()
									ev := StreamEvent(line)
									if ev.Valid() {
										// valid reply
										cmd := ev.Command()
										if cmd == RPL_StreamingTransfered {
											// successful transfer
											log.WithFields(log.Fields{
												"feed":  c.C.state.FeedName,
												"msgid": msgid,
												"bytes": n,
											}).Debug("Article Transferred")
											// call hooks
											if c.C.hooks != nil {
												c.C.hooks.SentArticleVia(msgid, c.C.state.FeedName)
											}
										} else {
											// failed transfer
											log.WithFields(log.Fields{
												"feed":  c.C.state.FeedName,
												"msgid": msgid,
												"bytes": n,
											}).Debug("Article Rejected")
										}
									}
								} else {
									log.WithFields(log.Fields{
										"feed":  c.C.state.FeedName,
										"msgid": msgid,
									}).Errorf("failed to transfer: %s", err.Error())
								}
							}
						} else {
							log.WithFields(log.Fields{
								"pkg":   "nntp-conn",
								"state": c.C.state,
								"msgid": msgid,
							}).Warn("article not in storage, not sending")
						}
					}
				} else {
					// invalid reply
					log.WithFields(log.Fields{
						"pkg":   "nntp-conn",
						"state": c.C.state,
						"msgid": msgid,
						"line":  line,
					}).Error("invalid streaming response")
					// close
					return
				}
			} else {
				log.WithFields(log.Fields{
					"pkg":   "nntp-conn",
					"state": c.C.state,
					"msgid": msgid,
				}).Error("streaming error during CHECK", err)
				return
			}
		} else {
			// channel closed
			return
		}
	}
}

func (c *v1OBConn) Quit() {
	c.C.printfLine("QUIT yo")
	c.C.readline()
	c.C.Close()
}

func (c *v1OBConn) StartStreaming() (chnl chan ArticleEntry, err error) {
	if c.streamChnl == nil {
		c.streamChnl = make(chan ArticleEntry)

	}
	chnl = c.streamChnl
	return
}

func (c *v1OBConn) GetState() *ConnState {
	return c.GetState()
}

// create a new connection from an established connection
func newOutboundConn(c net.Conn, s *Server, conf *config.FeedConfig) Conn {

	sname := s.Name()

	if len(sname) == 0 {
		sname = "nntp.anon.tld"
	}
	storage := s.Storage
	if storage == nil {
		storage = store.NewNullStorage()
	}
	return &v1OBConn{
		conf: conf,
		C: v1Conn{
			hooks: s,
			state: ConnState{
				FeedName: conf.Name,
				HostName: conf.Addr,
				Open:     true,
			},
			serverName: sname,
			storage:    storage,
			C:          textproto.NewConn(c),
			conn:       c,
			hdrio:      message.NewHeaderIO(),
		},
	}
}

type v1IBConn struct {
	C v1Conn
}

func (c *v1IBConn) DownloadGroup(g Newsgroup) error {
	return nil
}

func (c *v1IBConn) ListNewsgroups() (groups []Newsgroup, err error) {
	return
}

func (c *v1IBConn) GetState() *ConnState {
	return c.C.GetState()
}

// negotiate an inbound connection
func (c *v1IBConn) Negotiate(stream bool) (err error) {
	var line string
	if c.PostingAllowed() {
		line = Line_PostingAllowed
	} else {
		line = Line_PostingNotAllowed
	}
	err = c.C.printfLine(line)
	return
}

func (c *v1IBConn) PostingAllowed() bool {
	return c.C.PostingAllowed()
}

func (c *v1IBConn) IsOpen() bool {
	return c.C.IsOpen()
}

func (c *v1IBConn) Quit() {
	// inbound connections quit without warning
	log.WithFields(log.Fields{
		"pkg":  "nntp-ibconn",
		"addr": c.C.conn.RemoteAddr(),
	}).Info("closing inbound connection")
	c.C.Close()
}

// is this connection authenticated?
func (c *v1Conn) Authed() bool {
	return c.tlsConn != nil || c.authenticated
}

// unconditionally close connection
func (c *v1Conn) Close() {
	if c.tlsConn == nil {
		// tls is not on
		c.C.Close()
	} else {
		// tls is on
		// we should close tls cleanly
		c.tlsConn.Close()
	}
	c.state.Open = false
}

func (c *v1IBConn) WantsStreaming() bool {
	return c.C.state.Mode.Is(MODE_STREAM)
}

func (c *v1Conn) printfLine(format string, args ...interface{}) error {
	log.WithFields(log.Fields{
		"pkg":     "nntp-conn",
		"version": 1,
		"state":   &c.state,
		"io":      "send",
	}).Debugf(format, args...)
	return c.C.PrintfLine(format, args...)
}

func (c *v1Conn) readline() (line string, err error) {
	line, err = c.C.ReadLine()
	log.WithFields(log.Fields{
		"pkg":     "nntp-conn",
		"version": 1,
		"state":   &c.state,
		"io":      "recv",
	}).Debug(line)
	return
}

// handle switching nntp modes for inbound connection
func switchModeInbound(c *v1Conn, line string, hooks EventHooks) (err error) {
	cmd := ModeCommand(line)
	m := c.Mode()
	if cmd.Is(ModeReader) {
		if m.Is(MODE_STREAM) {
			// we need to stop streaming
		}
		var line string
		if c.PostingAllowed() {
			line = Line_PostingAllowed
		} else {
			line = Line_PostingNotAllowed
		}
		err = c.printfLine(line)
		if err == nil {
			c.state.Mode = MODE_READER
		}
	} else if cmd.Is(ModeStream) {
		// we want to switch to streaming mode
		err = c.printfLine(Line_StreamingAllowed)
		if err == nil {
			c.state.Mode = MODE_STREAM
		}
	} else {
		err = c.printfLine(Line_InvalidMode)
	}
	return
}

// handle quit command
func quitConnection(c *v1Conn, line string, hooks EventHooks) (err error) {
	log.WithFields(log.Fields{
		"pkg":     "nntp-conn",
		"version": "1",
		"state":   &c.state,
	}).Debug("quit requested")
	err = c.printfLine(Line_RPLQuit)
	c.Close()
	return
}

// send our capabailities
func sendCapabilities(c *v1Conn, line string, hooks EventHooks) (err error) {
	var caps []string

	caps = append(caps, "MODE-READER", "IMPLEMENTATION nntpchand", "STREAMING")
	if c.tlsConfig != nil {
		caps = append(caps, "STARTTLS")
	}

	err = c.printfLine("%s We can do things", RPL_Capabilities)
	if err == nil {
		for _, l := range caps {
			err = c.printfLine(l)
			if err != nil {
				log.WithFields(log.Fields{
					"pkg":     "nntp-conn",
					"version": "1",
					"state":   &c.state,
				}).Error(err)
			}
		}
		err = c.printfLine(".")
	}
	return
}

// read an article via dotreader
func (c *v1Conn) readArticle(newpost bool, hooks EventHooks) (ps PolicyStatus, err error) {
	store_r, store_w := io.Pipe()
	article_r, article_w := io.Pipe()
	article_body_r, article_body_w := io.Pipe()

	accept_chnl := make(chan PolicyStatus)
	store_info_chnl := make(chan ArticleEntry)
	store_result_chnl := make(chan error)

	hdr_chnl := make(chan message.Header)

	log.WithFields(log.Fields{
		"pkg": "nntp-conn",
	}).Debug("start reading")
	done_chnl := make(chan PolicyStatus)
	go func() {
		var err error
		dr := c.C.DotReader()
		var buff [1024]byte
		var n int64
		n, err = io.CopyBuffer(article_w, dr, buff[:])
		log.WithFields(log.Fields{
			"n": n,
		}).Debug("read from connection")
		if err != nil && err != io.EOF {
			article_w.CloseWithError(err)
		} else {
			article_w.Close()
		}
		st := <-accept_chnl
		close(accept_chnl)
		// get result from storage
		err2, ok := <-store_result_chnl
		if ok && err2 != io.EOF {
			err = err2
		}
		close(store_result_chnl)
		done_chnl <- st
	}()

	// parse message and store attachments in bg
	go func(msgbody io.ReadCloser) {
		defer msgbody.Close()
		hdr, ok := <-hdr_chnl
		if !ok {
			return
		}
		// all text in this post
		// txt := new(bytes.Buffer)
		// the article itself
		// a := new(model.Article)
		var err error
		if hdr.IsMultipart() {
			var params map[string]string
			_, params, err = hdr.GetMediaType()
			if err == nil {
				boundary, ok := params["boundary"]
				if ok {
					part_r := multipart.NewReader(msgbody, boundary)
					for err == nil {
						var part *multipart.Part
						part, err = part_r.NextPart()
						if err == io.EOF {
							// we done
							break
						} else if err == nil {
							// we gots a part

							// get header
							part_hdr := part.Header

							// check for base64 encoding
							var part_body io.Reader
							if part_hdr.Get("Content-Transfer-Encoding") == "base64" {
								part_body = base64.NewDecoder(base64.StdEncoding, part)
							} else {
								part_body = part
							}

							// get content type
							content_type := part_hdr.Get("Content-Type")
							if len(content_type) == 0 {
								// assume text/plain
								content_type = "text/plain; charset=UTF8"
							}
							var part_type string
							// extract mime type
							part_type, _, err = mime.ParseMediaType(content_type)
							if err == nil {

								if part_type == "text/plain" {
									// if we are plaintext save it to the text buffer
									_, err = io.Copy(util.Discard, part_body)
								} else {
									var fpath string
									fname := part.FileName()
									fpath, err = c.storage.StoreAttachment(part_body, fname)
									if err == nil {
										// stored attachment good
										log.WithFields(log.Fields{
											"pkg":      "nntp-conn",
											"state":    &c.state,
											"version":  "1",
											"filename": fname,
											"filepath": fpath,
										}).Debug("attachment stored")
									} else {
										// failed to save attachment
										log.WithFields(log.Fields{
											"pkg":     "nntp-conn",
											"state":   &c.state,
											"version": "1",
										}).Error("failed to save attachment ", err)
									}
								}
							} else {
								// cannot read part header
								log.WithFields(log.Fields{
									"pkg":     "nntp-conn",
									"state":   &c.state,
									"version": "1",
								}).Error("bad attachment in multipart message ", err)
							}
							err = nil
							part.Close()
						} else if err != io.EOF {
							// error reading part
							log.WithFields(log.Fields{
								"pkg":     "nntp-conn",
								"state":   &c.state,
								"version": "1",
							}).Error("error reading part ", err)
						}
					}
				}
			}
		} else if hdr.IsSigned() {
			// signed message

			// discard for now
			_, err = io.Copy(util.Discard, msgbody)
		} else {
			// plaintext message
			var n int64
			n, err = io.Copy(util.Discard, msgbody)
			log.WithFields(log.Fields{
				"bytes": n,
				"pkg":   "nntp-conn",
			}).Debug("text body copied")
		}
		if err != nil && err != io.EOF {
			log.WithFields(log.Fields{
				"pkg":   "nntp-conn",
				"state": &c.state,
			}).Error("error handing message body", err)
		}
	}(article_body_r)

	// store function
	go func(r io.ReadCloser) {
		e, ok := <-store_info_chnl
		if !ok {
			// failed to get info
			// don't read anything
			r.Close()
			store_result_chnl <- io.EOF
			return
		}
		msgid := e.MessageID()
		if msgid.Valid() {
			// valid message-id
			log.WithFields(log.Fields{
				"pkg":     "nntp-conn",
				"msgid":   msgid,
				"version": "1",
				"state":   &c.state,
			}).Debug("storing article")

			fpath, err := c.storage.StoreArticle(r, msgid.String(), e.Newsgroup().String())
			r.Close()
			if err == nil {
				log.WithFields(log.Fields{
					"pkg":     "nntp-conn",
					"msgid":   msgid,
					"version": "1",
					"state":   &c.state,
				}).Debug("stored article okay to ", fpath)
				// we got the article
				if hooks != nil {
					hooks.GotArticle(msgid, e.Newsgroup())
				}
				store_result_chnl <- io.EOF
				log.Debugf("store informed")
			} else {
				// error storing article
				log.WithFields(log.Fields{
					"pkg":     "nntp-conn",
					"msgid":   msgid,
					"state":   &c.state,
					"version": "1",
				}).Error("failed to store article ", err)
				io.Copy(util.Discard, r)
				store_result_chnl <- err
			}
		} else {
			// invalid message-id
			// discard
			log.WithFields(log.Fields{
				"pkg":     "nntp-conn",
				"msgid":   msgid,
				"state":   &c.state,
				"version": "1",
			}).Warn("store will discard message with invalid message-id")
			io.Copy(util.Discard, r)
			store_result_chnl <- nil
			r.Close()
		}
	}(store_r)

	// acceptor function
	go func(r io.ReadCloser, out_w, body_w io.WriteCloser) {
		var w io.WriteCloser
		defer r.Close()
		status := PolicyAccept
		hdr, err := c.hdrio.ReadHeader(r)
		if err == nil {
			// append path
			hdr.AppendPath(c.serverName)
			// get message-id
			var msgid MessageID
			if newpost {
				// new post
				// generate it
				msgid = GenMessageID(c.serverName)
				hdr.Set("Message-ID", msgid.String())
			} else {
				// not a new post, get from header
				msgid = MessageID(hdr.MessageID())
				if msgid.Valid() {
					// check store fo existing article
					err = c.storage.HasArticle(msgid.String())
					if err == store.ErrNoSuchArticle {
						// we don't have the article
						status = PolicyAccept
						log.Infof("accept article %s", msgid)
					} else if err == nil {
						// we do have the article, reject it we don't need it again
						status = PolicyReject
					} else {
						// some other error happened
						log.WithFields(log.Fields{
							"pkg":   "nntp-conn",
							"state": c.state,
						}).Error("failed to check store for article ", err)
					}
					err = nil
				} else {
					// bad article
					status = PolicyBan
				}
			}
			// check the header if we have an acceptor and the previous checks are good
			if status.Accept() && c.acceptor != nil {
				status = c.acceptor.CheckHeader(hdr)
			}
			if status.Accept() {
				// we have accepted the article
				// store to disk
				w = out_w
			} else {
				// we have not accepted the article
				// discard
				w = util.Discard
				out_w.Close()
			}
			store_info_chnl <- ArticleEntry{msgid.String(), hdr.Newsgroup()}
			hdr_chnl <- hdr
			// close the channel for headers
			close(hdr_chnl)
			// write header out to storage
			err = c.hdrio.WriteHeader(hdr, w)
			if err == nil {
				mw := io.MultiWriter(body_w, w)
				// we wrote header
				var n int64
				if c.acceptor == nil {
					// write the rest of the body
					// we don't care about article size
					log.WithFields(log.Fields{}).Debug("copying body")
					var buff [128]byte
					n, err = io.CopyBuffer(mw, r, buff[:])
				} else {
					// we care about the article size
					max := c.acceptor.MaxArticleSize()
					var n int64
					// copy it out
					n, err = io.CopyN(mw, r, max)
					if err == nil {
						if n < max {
							// under size limit
							// we gud
							log.WithFields(log.Fields{
								"pkg":   "nntp-conn",
								"bytes": n,
								"state": &c.state,
							}).Debug("body fits")
						} else {
							// too big, discard the rest
							_, err = io.Copy(util.Discard, r)
							// ... and ban it
							status = PolicyBan
						}
					}
				}
				log.WithFields(log.Fields{
					"pkg":   "nntp-conn",
					"bytes": n,
					"state": &c.state,
				}).Debug("body wrote")
				// TODO: inform store to delete article and attachments
			} else {
				// error writing header
				log.WithFields(log.Fields{
					"msgid": msgid,
				}).Error("error writing header ", err)
			}
		} else {
			// error reading header
			// possibly a read error?
			status = PolicyDefer
		}
		// close info channel for store
		close(store_info_chnl)
		w.Close()
		// close body pipe
		body_w.Close()
		// inform result
		log.Debugf("status %s", status)
		accept_chnl <- status
		log.Debugf("informed")
	}(article_r, store_w, article_body_w)

	ps = <-done_chnl
	close(done_chnl)
	log.Debug("read article done")
	return
}

// handle IHAVE command
func nntpRecvArticle(c *v1Conn, line string, hooks EventHooks) (err error) {
	parts := strings.Split(line, " ")
	if len(parts) == 2 {
		msgid := MessageID(parts[1])
		if msgid.Valid() {
			// valid message-id
			err = c.printfLine("%s send article to be transfered", RPL_TransferAccepted)
			// read in article
			if err == nil {
				var status PolicyStatus
				status, err = c.readArticle(false, hooks)
				if err == nil {
					// we read in article
					if status.Accept() {
						// accepted
						err = c.printfLine("%s transfer wuz gud", RPL_TransferOkay)
					} else if status.Defer() {
						// deferred
						err = c.printfLine("%s transfer defer", RPL_TransferDefer)
					} else if status.Reject() {
						// rejected
						err = c.printfLine("%s transfer rejected, don't send it again brah", RPL_TransferReject)
					}
				} else {
					// could not transfer article
					err = c.printfLine("%s transfer failed; try again later", RPL_TransferDefer)
				}
			}
		} else {
			// invalid message-id
			err = c.printfLine("%s article not wanted", RPL_TransferNotWanted)
		}
	} else {
		// invaldi syntax
		err = c.printfLine("%s invalid syntax", RPL_SyntaxError)
	}
	return
}

// handle POST command
func nntpPostArticle(c *v1Conn, line string, hooks EventHooks) (err error) {
	if c.PostingAllowed() {
		if c.Mode().Is(MODE_READER) {
			err = c.printfLine("%s go ahead yo", RPL_PostAccepted)
			var status PolicyStatus
			status, err = c.readArticle(true, hooks)
			if err == nil {
				// read okay
				if status.Accept() {
					err = c.printfLine("%s post was recieved", RPL_PostReceived)
				} else {
					err = c.printfLine("%s posting failed", RPL_PostingFailed)
				}
			} else {
				log.WithFields(log.Fields{
					"pkg":     "nntp-conn",
					"state":   &c.state,
					"version": "1",
				}).Error("POST failed ", err)
				err = c.printfLine("%s post failed: %s", RPL_PostingFailed, err)
			}
		} else {
			// not in reader mode
			err = c.printfLine("%s not in reader mode", RPL_WrongMode)
		}
	} else {
		err = c.printfLine("%s posting is disallowed", RPL_PostingNotPermitted)
	}
	return
}

// handle streaming line
func streamingLine(c *v1Conn, line string, hooks EventHooks) (err error) {
	ev := StreamEvent(line)
	if c.Mode().Is(MODE_STREAM) {
		if ev.Valid() {
			// valid stream line
			cmd := ev.Command()
			msgid := ev.MessageID()
			if cmd == stream_CHECK {
				if c.acceptor == nil {
					// no acceptor, we'll take them all
					err = c.printfLine("%s %s", RPL_StreamingAccept, msgid)
				} else {
					status := PolicyAccept
					if c.storage.HasArticle(msgid.String()) == nil {
						// we have this article
						status = PolicyReject
					}
					if status.Accept() && c.acceptor != nil {
						status = c.acceptor.CheckMessageID(ev.MessageID())
					}
					if status.Accept() {
						// accepted
						err = c.printfLine("%s %s", RPL_StreamingAccept, msgid)
					} else if status.Defer() {
						// deferred
						err = c.printfLine("%s %s", RPL_StreamingDefer, msgid)
					} else {
						// rejected
						err = c.printfLine("%s %s", RPL_StreamingReject, msgid)
					}
				}
			} else if cmd == stream_TAKETHIS {
				var status PolicyStatus
				status, err = c.readArticle(false, hooks)
				if status.Accept() {
					// this article was accepted
					err = c.printfLine("%s %s", RPL_StreamingTransfered, msgid)
				} else {
					// this article was not accepted
					err = c.printfLine("%s %s", RPL_StreamingReject, msgid)
				}
			}
		} else {
			// invalid line
			err = c.printfLine("%s Invalid syntax", RPL_SyntaxError)
		}
	} else {
		if ev.MessageID().Valid() {
			// not in streaming mode
			err = c.printfLine("%s %s", RPL_StreamingDefer, ev.MessageID())
		} else {
			// invalid message id
			err = c.printfLine("%s Invalid Syntax", RPL_SyntaxError)
		}
	}
	return
}

func newsgroupList(c *v1Conn, line string, hooks EventHooks, rpl string) (err error) {
	var groups []string
	if c.storage == nil {
		// no database driver available
		// let's say we carry overchan.test for now
		groups = append(groups, "overchan.test")
	} else {
		groups, err = c.storage.GetAllNewsgroups()
	}

	if err == nil {
		// we got newsgroups from the db
		dw := c.C.DotWriter()
		fmt.Fprintf(dw, "%s list of newsgroups follows\n", rpl)
		for _, g := range groups {
			hi := uint64(1)
			lo := uint64(0)
			if c.storage != nil {
				hi, lo, err = c.storage.GetWatermark(g)
			}
			if err != nil {
				// log error if it occurs
				log.WithFields(log.Fields{
					"pkg":   "nntp-conn",
					"group": g,
					"state": c.state,
				}).Warn("cannot get high low water marks for LIST command")

			} else {
				fmt.Fprintf(dw, "%s %d %d y", g, hi, lo)
			}
		}
		// flush dotwriter
		err = dw.Close()
	} else {
		// db error while getting newsgroup list
		err = c.printfLine("%s cannot list newsgroups %s", RPL_GenericError, err.Error())
	}
	return
}

// handle inbound STARTTLS command
func upgradeTLS(c *v1Conn, line string, hooks EventHooks) (err error) {
	if c.tlsConfig == nil {
		err = c.printfLine("%s TLS not supported", RPL_TLSRejected)
	} else {
		err = c.printfLine("%s Continue with TLS Negotiation", RPL_TLSContinue)
		if err == nil {
			tconn := tls.Server(c.conn, c.tlsConfig)
			err = tconn.Handshake()
			if err == nil {
				// successful tls handshake
				c.tlsConn = tconn
				c.C = textproto.NewConn(c.tlsConn)
			} else {
				// tls failed
				log.WithFields(log.Fields{
					"pkg":   "nntp-conn",
					"addr":  c.conn.RemoteAddr(),
					"state": c.state,
				}).Warn("TLS Handshake failed ", err)
				// fall back to plaintext
				err = nil
			}
		}
	}
	return
}

// switch to another newsgroup
func switchNewsgroup(c *v1Conn, line string, hooks EventHooks) (err error) {
	parts := strings.Split(line, " ")
	var has bool
	var group Newsgroup
	if len(parts) == 2 {
		group = Newsgroup(parts[1])
		if group.Valid() {
			// correct format
			if c.storage == nil {
				has = false
			} else {
				has, err = c.storage.HasNewsgroup(group.String())
			}
		}
	}
	if has {
		// we have it
		hi := uint64(1)
		lo := uint64(0)
		if c.storage != nil {
			// check database for water marks
			hi, lo, err = c.storage.GetWatermark(group.String())
		}
		if err == nil {
			// XXX: ensure hi > lo
			err = c.printfLine("%s %d %d %d %s", RPL_Group, hi-lo, lo, hi, group.String())
			if err == nil {
				// line was sent
				c.state.Group = group
				log.WithFields(log.Fields{
					"pkg":   "nntp-conn",
					"group": group,
					"state": c.state,
				}).Debug("switched newsgroups")
			}
		} else {
			err = c.printfLine("%s error checking for newsgroup %s", RPL_GenericError, err.Error())
		}
	} else if err != nil {
		// error
		err = c.printfLine("%s error checking for newsgroup %s", RPL_GenericError, err.Error())
	} else {
		// incorrect format
		err = c.printfLine("%s no such newsgroup", RPL_NoSuchGroup)
	}
	return
}

func handleAuthInfo(c *v1Conn, line string, hooks EventHooks) (err error) {
	subcmd := line[9:]
	if strings.HasPrefix(strings.ToUpper(subcmd), "USER") {
		c.username = subcmd[5:]
		err = c.printfLine("%s password required", RPL_MoreAuth)
	} else if strings.HasPrefix(strings.ToUpper(subcmd), "PASS") {
		var success bool
		if c.username == "" {
			// out of order commands
			c.printfLine("%s auth info sent out of order yo", RPL_GenericError)
			return
		} else if c.auth == nil {
			// no auth mechanism, this will be set to true if anon nntp is enabled
			success = c.authenticated
		} else {
			// check login
			success, err = c.auth.CheckLogin(c.username, subcmd[5:])
		}
		if success {
			// login good
			err = c.printfLine("%s login gud, proceed yo", RPL_AuthAccepted)
			c.authenticated = true
		} else if err == nil {
			// login bad
			err = c.printfLine("%s bad login", RPL_AuthenticateRejected)
		} else {
			// error
			err = c.printfLine("%s error processing login: %s", RPL_GenericError, err.Error())
		}
	} else {
		err = c.printfLine("%s only USER/PASS accepted with AUTHINFO", RPL_SyntaxError)
	}
	return
}

func handleXOVER(c *v1Conn, line string, hooks EventHooks) (err error) {
	group := c.Group()
	if group == "" {
		err = c.printfLine("%s no group selected", RPL_NoGroupSelected)
		return
	}
	if !Newsgroup(group).Valid() {
		err = c.printfLine("%s Invalid Newsgroup format", RPL_GenericError)
		return
	}
	err = c.printfLine("%s overview follows", RPL_Overview)
	if err != nil {
		return
	}
	chnl := make(chan string)
	go func() {
		c.storage.ForEachInGroup(group, chnl)
		close(chnl)
	}()
	i := 0
	for err == nil {
		m, ok := <-chnl
		if !ok {
			break
		}
		msgid := MessageID(m)
		if !msgid.Valid() {
			continue
		}
		var f *os.File
		f, err = c.storage.OpenArticle(m)
		if f != nil {
			h, e := c.hdrio.ReadHeader(f)
			f.Close()
			if e == nil {
				i++
				err = c.printfLine("%.6d\t%s\t%s\t%s\t%s\t%s", i, h.Get("Subject", "None"), h.Get("From", "anon <anon@anon.tld>"), h.Get("Date", "???"), h.MessageID(), h.Reference())
			}
		}
	}
	if err == nil {
		err = c.printfLine(".")
	}
	return
}

func handleArticle(c *v1Conn, line string, hooks EventHooks) (err error) {
	msgid := MessageID(line[8:])
	if msgid.Valid() && c.storage.HasArticle(msgid.String()) == nil {
		// valid id and we have it
		var r io.ReadCloser
		var buff [1024]byte
		r, err = c.storage.OpenArticle(msgid.String())
		if err == nil {
			err = c.printfLine("%s %s", RPL_Article, msgid)
			for err == nil {
				_, err = io.CopyBuffer(c.C.W, r, buff[:])
			}
			if err == io.EOF {
				err = nil
			}
			if err == nil {
				err = c.printfLine(".")
			}
			r.Close()
			return
		}
	}
	// invalid id or we don't have it
	err = c.printfLine("%s %s", RPL_NoArticleMsgID, msgid)
	return
}

// inbound streaming start
func (c *v1IBConn) StartStreaming() (chnl chan ArticleEntry, err error) {
	if c.Mode().Is(MODE_STREAM) {
		chnl = make(chan ArticleEntry)
	} else {
		err = ErrInvalidMode
	}
	return
}

func (c *v1IBConn) Mode() Mode {
	return c.C.Mode()
}

func (c *v1IBConn) ProcessInbound(hooks EventHooks) {
	c.C.Process(hooks)
}

// inbound streaming handling
func (c *v1IBConn) StreamAndQuit() {
}

func newInboundConn(s *Server, c net.Conn) Conn {
	sname := s.Name()
	storage := s.Storage
	if storage == nil {
		storage = store.NewNullStorage()
	}
	anon := false
	if s.Config != nil {
		anon = s.Config.AnonNNTP
	}
	return &v1IBConn{
		C: v1Conn{
			state: ConnState{
				FeedName: "inbound-feed",
				HostName: c.RemoteAddr().String(),
				Open:     true,
			},
			auth:          s.Auth,
			authenticated: anon,
			serverName:    sname,
			storage:       storage,
			acceptor:      s.Acceptor,
			hdrio:         message.NewHeaderIO(),
			C:             textproto.NewConn(c),
			conn:          c,
			cmds: map[string]lineHandlerFunc{
				"STARTTLS":     upgradeTLS,
				"IHAVE":        nntpRecvArticle,
				"POST":         nntpPostArticle,
				"MODE":         switchModeInbound,
				"QUIT":         quitConnection,
				"CAPABILITIES": sendCapabilities,
				"CHECK":        streamingLine,
				"TAKETHIS":     streamingLine,
				"LIST": func(c *v1Conn, line string, h EventHooks) error {
					return newsgroupList(c, line, h, RPL_List)
				},
				"NEWSGROUPS": func(c *v1Conn, line string, h EventHooks) error {
					return newsgroupList(c, line, h, RPL_NewsgroupList)
				},
				"GROUP":    switchNewsgroup,
				"AUTHINFO": handleAuthInfo,
				"XOVER":    handleXOVER,
				"ARTICLE":  handleArticle,
			},
		},
	}
}
