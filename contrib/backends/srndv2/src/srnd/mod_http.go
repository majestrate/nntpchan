//
// mod_http.go
//
// http mod panel
//

package srnd

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/csrf"
	"github.com/gorilla/sessions"
	"github.com/majestrate/nacl"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
)

type httpModUI struct {
	regenAll         func()
	regen            func(ArticleEntry)
	regenGroup       func(string)
	delete           func(string)
	deleteBoardPages func(string)
	modMessageChan   chan NNTPMessage
	daemon           *NNTPDaemon
	articles         ArticleStore
	store            *sessions.CookieStore
	prefix           string
	mod_prefix       string
}

func createHttpModUI(frontend *httpFrontend) httpModUI {
	return httpModUI{frontend.regenAll, frontend.Regen, frontend.regenerateBoard, frontend.deleteThreadMarkup, frontend.deleteBoardMarkup, make(chan NNTPMessage), frontend.daemon, frontend.daemon.store, frontend.store, frontend.prefix, frontend.prefix + "mod/"}

}

func extractGroup(param map[string]interface{}) string {
	return extractParam(param, "newsgroup")
}

func (self httpModUI) getAdminFunc(funcname string) AdminFunc {
	if funcname == "template.reload" {
		return func(param map[string]interface{}) (interface{}, error) {
			tname, ok := param["template"]
			if ok {
				t := ""
				switch tname.(type) {
				case string:
					t = tname.(string)
				default:
					return "failed to reload templates", errors.New("invalid parameters")
				}
				template.reloadTemplate(t)
				return "reloaded " + t, nil
			}
			template.reloadAllTemplates()
			return "reloaded all templates", nil
		}
	} else if funcname == "frontend.regen" {
		return func(param map[string]interface{}) (interface{}, error) {
			newsgroup := extractGroup(param)
			if len(newsgroup) > 0 {
				if self.daemon.database.HasNewsgroup(newsgroup) {
					go self.regenGroup(newsgroup)
				} else {
					return "failed to regen group", errors.New("no such board")
				}
			} else {
				go self.regenAll()
			}
			return "started regeneration", nil
		}
	} else if funcname == "thumbnail.regen" {
		return func(param map[string]interface{}) (interface{}, error) {
			threads, ok := param["threads"]
			t := 1
			if ok {
				switch threads.(type) {
				case int64:
					t = int(threads.(int64))
					if t <= 0 {
						return "failed to regen thumbnails", errors.New("invalid number of threads")
					}
				default:
					return "failed to regen thumbnails", errors.New("invalid parameters")
				}
			}
			log.Println("regenerating all thumbnails with", t, "threads")
			msgid := extractParam(param, "message-id")
			if ValidMessageID(msgid) {
				a := self.daemon.database.GetPostAttachments(msgid)
				go func(atts []string) {
					for _, att := range atts {
						self.articles.GenerateThumbnail(att)
					}
				}(a)
				return fmt.Sprintf("regenerating %d thumbnails for %s", len(a), msgid), nil
			}
			go reThumbnail(t, self.articles, true)
			return fmt.Sprintf("started rethumbnailing with %d threads", t), nil
		}
	} else if funcname == "frontend.add" {
		return func(param map[string]interface{}) (interface{}, error) {
			newsgroup := extractGroup(param)
			if len(newsgroup) > 0 && newsgroupValidFormat(newsgroup) && strings.HasPrefix(newsgroup, "overchan.") && newsgroup != "overchan." {
				if self.daemon.database.HasNewsgroup(newsgroup) {
					// we already have this newsgroup
					return "already have that newsgroup", nil
				} else {
					// we dont got this newsgroup
					log.Println("adding newsgroup", newsgroup)
					self.daemon.database.RegisterNewsgroup(newsgroup)
					return "added " + newsgroup, nil
				}
			}
			return "bad newsgroup", errors.New("invalid newsgroup name: " + newsgroup)
		}
	} else if funcname == "frontend.ban" {
		return func(param map[string]interface{}) (interface{}, error) {
			newsgroup := extractGroup(param)
			if len(newsgroup) > 0 {
				log.Println("banning", newsgroup)
				// check ban
				banned, err := self.daemon.database.NewsgroupBanned(newsgroup)
				if banned {
					// already banned
					return "cannot ban newsgroup", errors.New("already banned " + newsgroup)
				} else if err == nil {
					// do the ban here
					err = self.daemon.database.BanNewsgroup(newsgroup)
					// check for error
					if err == nil {
						// all gud
						return "banned " + newsgroup, nil
					} else {
						// error while banning
						return "error banning newsgroup", err
					}
				} else {
					// error checking ban
					return "cannot check ban", err
				}
			} else {
				// bad parameters
				return "cannot ban newsgroup", errors.New("invalid parameters")
			}
		}
	} else if funcname == "frontend.unban" {
		return func(param map[string]interface{}) (interface{}, error) {
			newsgroup := extractGroup(param)
			if len(newsgroup) > 0 {
				log.Println("unbanning", newsgroup)
				err := self.daemon.database.UnbanNewsgroup(newsgroup)
				if err == nil {
					return "unbanned " + newsgroup, nil
				} else {
					return "couldn't unban " + newsgroup, err
				}
			} else {
				return "cannot unban", errors.New("invalid paramters")
			}
		}
	} else if funcname == "frontend.nuke" {
		return func(param map[string]interface{}) (interface{}, error) {
			newsgroup := extractGroup(param)
			if len(newsgroup) > 0 {
				log.Println("nuking", newsgroup)
				// get every thread we have in this group
				for _, entry := range self.daemon.database.GetLastBumpedThreads(newsgroup, 10000) {
					// delete their thread page
					self.delete(entry.MessageID())
				}
				// delete every board page
				self.deleteBoardPages(newsgroup)
				go self.daemon.database.NukeNewsgroup(newsgroup, self.articles)
				return "nuke started", nil
			} else {
				return "cannot nuke", errors.New("invalid parameters")
			}
		}
	} else if funcname == "pubkey.add" {
		return func(param map[string]interface{}) (interface{}, error) {
			pubkey := extractParam(param, "pubkey")
			group := extractGroup(param)
			if group == "" {
				log.Println("pubkey.add global mod", pubkey)
				if self.daemon.database.CheckModPubkeyGlobal(pubkey) {
					return "already added", nil
				} else {
					err := self.daemon.database.MarkModPubkeyGlobal(pubkey)
					if err == nil {
						return "added", nil
					} else {
						return "error", err
					}
				}
			} else if newsgroupValidFormat(group) {
				log.Println("pubkey.add", group, "mod", pubkey)
				if self.daemon.database.CheckModPubkeyCanModGroup(pubkey, group) {
					return "already added", nil
				}
				err := self.daemon.database.MarkModPubkeyCanModGroup(pubkey, group)
				if err == nil {
					return "added", nil
				} else {
					return "error", err
				}
			} else {
				return "bad newsgroup: " + group, nil
			}
		}
	} else if funcname == "pubkey.del" {
		return func(param map[string]interface{}) (interface{}, error) {
			pubkey := extractParam(param, "pubkey")
			log.Println("pubkey.del", pubkey)
			if self.daemon.database.CheckModPubkeyGlobal(pubkey) {
				err := self.daemon.database.UnMarkModPubkeyGlobal(pubkey)
				if err == nil {
					return "removed", nil
				} else {
					return "error", err
				}
			} else {
				return "key not already trusted", nil
			}
		}

	} else if funcname == "nntp.login.del" {
		return func(param map[string]interface{}) (interface{}, error) {
			username := extractParam(param, "username")
			if len(username) > 0 {
				exists, err := self.daemon.database.CheckNNTPUserExists(username)
				if exists {
					err = self.daemon.database.RemoveNNTPLogin(username)
					if err == nil {
						return "removed user", nil
					}
					return "", nil
				} else if err == nil {
					return "no such user", nil
				} else {
					return "", err
				}
			} else {
				return "no such user", nil
			}
		}
	} else if funcname == "nntp.login.add" {
		return func(param map[string]interface{}) (interface{}, error) {
			username := extractParam(param, "username")
			passwd := extractParam(param, "passwd")
			if len(username) > 0 && len(passwd) > 0 {
				log.Println("nntp.login.add", username)
				// check if users is there
				exists, err := self.daemon.database.CheckNNTPUserExists(username)
				if exists {
					// user is already there
					return "nntp user already exists", nil
				} else if err == nil {
					// now add the user
					err = self.daemon.database.AddNNTPLogin(username, passwd)
					// success adding?
					if err == nil {
						// yeh
						return "added nntp user", nil
					}
					// nah
					return "", err
				} else {
					// error happened
					return "", err
				}
			} else {
				return "invalid username or password format", nil
			}
		}
	} else if funcname == "feed.add" {
		return func(param map[string]interface{}) (interface{}, error) {
			host := extractParam(param, "host")
			port := extractParam(param, "port")
			name := extractParam(param, "name")
			if len(host) == 0 || len(port) == 0 || len(name) == 0 {
				// bad parameter
				return "", errors.New("please specific host, port and name")
			}
			// make new config
			conf := FeedConfig{
				policy: FeedPolicy{
					// default rules for default policy
					rules: map[string]string{"overchan.*": "0", "ctl": "1"},
				},
				Addr:   host + ":" + port,
				Name:   name,
				quarks: make(map[string]string),
			}
			err := self.daemon.addFeed(conf)
			if err == nil {
				return "feed added", err
			} else {
				return "", err
			}
		}
	} else if funcname == "feed.list" {
		return func(_ map[string]interface{}) (interface{}, error) {
			feeds := self.daemon.activeFeeds()
			return feeds, nil
		}
	} else if funcname == "feed.sync" {
		return func(_ map[string]interface{}) (interface{}, error) {
			go self.daemon.syncAllMessages()
			return "sync started", nil
		}
	} else if funcname == "feed.del" {
		return func(param map[string]interface{}) (interface{}, error) {
			name := extractParam(param, "name")
			self.daemon.removeFeed(name)
			return "okay", nil
		}
	} else if funcname == "store.expire" {
		return func(_ map[string]interface{}) (interface{}, error) {
			if self.daemon.expire == nil {
				// TODO: expire orphans?
				return "archive mode enabled, will not expire orphans", nil
			} else {
				go self.daemon.expire.ExpireOrphans()
				return "expiration started", nil
			}
		}
	} else if funcname == "frontend.posts" {
		// get all posts given parameters
		return func(param map[string]interface{}) (interface{}, error) {
			// by cidr
			cidr := extractParam(param, "cidr")
			// by encrypted ip
			encip := extractParam(param, "encip")

			var err error
			var post_msgids []string
			if len(cidr) > 0 {
				var cnet *net.IPNet
				_, cnet, err = net.ParseCIDR(cidr)
				if err == nil {
					post_msgids, err = self.daemon.database.GetMessageIDByCIDR(cnet)
				}
			} else if len(encip) > 0 {
				post_msgids, err = self.daemon.database.GetMessageIDByEncryptedIP(encip)
			}
			return post_msgids, err
		}
	}
	return nil
}

// handle an admin action
func (self httpModUI) HandleAdminCommand(wr http.ResponseWriter, r *http.Request) {
	self.asAuthed("admin", func(url string) {
		action := strings.Split(url, "/admin/")[1]
		f := self.getAdminFunc(action)
		if f == nil {
			wr.WriteHeader(404)
		} else {
			var result interface{}
			var err error
			req := make(map[string]interface{})
			if r.Method == "POST" {
				dec := json.NewDecoder(r.Body)
				err = dec.Decode(&req)
				r.Body.Close()
			}
			if err == nil {
				result, err = f(req)
			}
			resp := make(map[string]interface{})
			if err == nil {
				resp["error"] = nil
			} else {
				resp["error"] = err.Error()
			}
			resp["result"] = result
			enc := json.NewEncoder(wr)
			enc.Encode(resp)
		}

	}, wr, r)
}

func (self httpModUI) CheckPubkey(pubkey, scope string) (bool, error) {
	is_admin, err := self.daemon.database.CheckAdminPubkey(pubkey)
	if is_admin {
		// admin can do what they want
		return true, nil
	}
	if self.daemon.database.CheckModPubkeyGlobal(pubkey) {
		// this user is a global mod, can't do admin
		return scope != "admin", nil
	}
	// check for board specific mods
	if strings.Index(scope, "mod-") == 0 {
		group := scope[4:]
		if self.daemon.database.CheckModPubkeyCanModGroup(pubkey, group) {
			return true, nil
		}
	} else if scope == "login" {
		// check if a user can log in
		return self.daemon.database.CheckModPubkey(pubkey), nil
	}
	return false, err
}

func (self httpModUI) CheckKey(privkey, scope string) (bool, error) {
	privkey_bytes, err := hex.DecodeString(privkey)
	if err == nil {
		kp := nacl.LoadSignKey(privkey_bytes)
		if kp != nil {
			defer kp.Free()
			pubkey := hex.EncodeToString(kp.Public())
			return self.CheckPubkey(pubkey, scope)
		}
	}
	log.Println("invalid key format for key", privkey)
	return false, err
}

func (self httpModUI) MessageChan() chan NNTPMessage {
	return self.modMessageChan
}

func (self httpModUI) getSession(r *http.Request) *sessions.Session {
	s, _ := self.store.Get(r, "nntpchan-mod")
	return s
}

// get the session's private key as bytes or nil if we don't have it
func (self httpModUI) getSessionPrivkeyBytes(r *http.Request) []byte {
	s := self.getSession(r)
	k, ok := s.Values["privkey"]
	if ok {
		privkey_bytes, err := hex.DecodeString(k.(string))
		if err == nil {
			return privkey_bytes
		}
		log.Println("failed to decode private key bytes from session", err)
	} else {
		log.Println("failed to get private key from session, no private key in session?")
	}
	return nil
}

// returns true if the session is okay for a scope
// otherwise redirect to login page
func (self httpModUI) checkSession(r *http.Request, scope string) bool {
	s := self.getSession(r)
	k, ok := s.Values["privkey"]
	if ok {
		ok, err := self.CheckKey(k.(string), scope)
		if err != nil {
			return false
		}
		return ok
	}
	return false
}

func (self httpModUI) writeTemplate(wr http.ResponseWriter, r *http.Request, name string) {
	self.writeTemplateParam(wr, r, name, nil)
}

func (self httpModUI) writeTemplateParam(wr http.ResponseWriter, r *http.Request, name string, param map[string]interface{}) {
	if param == nil {
		param = make(map[string]interface{})
	}
	param[csrf.TemplateTag] = csrf.TemplateField(r)
	param["prefix"] = self.prefix
	param["mod_prefix"] = self.mod_prefix
	io.WriteString(wr, template.renderTemplate(name, param))
}

// do a function as authenticated
// pass in the request path to the handler
func (self httpModUI) asAuthed(scope string, handler func(string), wr http.ResponseWriter, r *http.Request) {
	if self.checkSession(r, scope) {
		handler(r.URL.Path)
	} else {
		wr.WriteHeader(403)
	}
}

// do stuff to a certain message if with have it and are authed
func (self httpModUI) asAuthedWithMessage(scope string, handler func(ArticleEntry, *http.Request) map[string]interface{}, wr http.ResponseWriter, req *http.Request) {
	self.asAuthed(scope, func(path string) {
		// get the long hash
		if strings.Count(path, "/") > 2 {
			// TOOD: prefix detection
			longhash := strings.Split(path, "/")[3]
			// get the message id
			msg, err := self.daemon.database.GetMessageIDByHash(longhash)
			resp := make(map[string]interface{})
			if err == nil {
				group := msg.Newsgroup()
				if err == nil {
					if self.checkSession(req, "mod-"+group) {
						// we can moderate this group
						resp = handler(msg, req)
					} else {
						// no permission to moderate this group
						resp["error"] = fmt.Sprint("you don't have permission to moderate '%s'", group)
					}
				} else {
					resp["error"] = err.Error()
				}
			} else {
				resp["error"] = fmt.Sprint("don't have post %s, %s", longhash, err.Error())
			}
			enc := json.NewEncoder(wr)
			enc.Encode(resp)
		} else {
			wr.WriteHeader(404)
		}
	}, wr, req)
}

func (self httpModUI) HandleAddPubkey(wr http.ResponseWriter, r *http.Request) {
}

func (self httpModUI) HandleDelPubkey(wr http.ResponseWriter, r *http.Request) {
}

func (self httpModUI) HandleUnbanAddress(wr http.ResponseWriter, r *http.Request) {
	self.asAuthed("ban", func(path string) {
		// extract the ip address
		// TODO: ip ranges and prefix detection
		if strings.Count(path, "/") > 2 {
			addr := strings.Split(path, "/")[3]
			resp := make(map[string]interface{})
			banned, err := self.daemon.database.CheckIPBanned(addr)
			if err != nil {
				resp["error"] = fmt.Sprintf("cannot tell if %s is banned: %s", addr, err.Error())
			} else if banned {
				// TODO: rangebans
				err = self.daemon.database.UnbanAddr(addr)
				if err == nil {
					resp["result"] = fmt.Sprintf("%s was unbanned", addr)
				} else {
					resp["error"] = err.Error()
				}
			} else {
				resp["error"] = fmt.Sprintf("%s was not banned", addr)
			}
			enc := json.NewEncoder(wr)
			enc.Encode(resp)
		} else {
			wr.WriteHeader(404)
		}
	}, wr, r)
}

// handle ban logic
func (self httpModUI) handleBanAddress(msg ArticleEntry, r *http.Request) map[string]interface{} {
	// get the article headers
	resp := make(map[string]interface{})
	msgid := msg.MessageID()
	hdr, err := self.daemon.database.GetHeadersForMessage(msgid)
	if hdr == nil {
		// we don't got it?!
		resp["error"] = fmt.Sprintf("could not load headers for %s: %s", msgid, err.Error())
	} else {
		// get the associated encrypted ip
		encip := hdr.Get("x-encrypted-ip", "")
		encip = strings.Trim(encip, "\t ")

		if len(encip) == 0 {
			// no ip header detected
			resp["error"] = fmt.Sprintf("%s has no IP, ban Tor instead", msgid)
		} else {
			// get the ip address if we have it
			ip, err := self.daemon.database.GetIPAddress(encip)
			if len(ip) > 0 {
				// we have it
				// ban the address
				err = self.daemon.database.BanAddr(ip)
				// then we tell everyone about it
				var key string
				// TODO: we SHOULD have the key, but what if we do not?
				key, err = self.daemon.database.GetEncKey(encip)
				// create mod message
				// TODO: hardcoded ban period
				mm := ModMessage{overchanInetBan(encip, key, -1)}
				privkey_bytes := self.getSessionPrivkeyBytes(r)
				if privkey_bytes == nil {
					// this should not happen
					log.Println("failed to get privkey bytes from session")
					resp["error"] = "failed to get private key from session. wtf?"
				} else {
					// wrap and sign
					nntp := wrapModMessage(mm)
					nntp, err = signArticle(nntp, privkey_bytes)
					if err == nil {
						// federate
						self.modMessageChan <- nntp
					}
				}
			} else {
				// we don't have it
				// ban the encrypted version
				err = self.daemon.database.BanEncAddr(encip)
			}
			if err == nil {
				result_msg := fmt.Sprintf("We banned %s", encip)
				if len(ip) > 0 {
					result_msg += fmt.Sprintf(" (%s)", ip)
				}
				resp["banned"] = result_msg
			} else {
				resp["error"] = err.Error()
			}

		}
	}
	return resp
}

func (self httpModUI) handleDeletePost(msg ArticleEntry, r *http.Request) map[string]interface{} {
	var mm ModMessage
	resp := make(map[string]interface{})
	msgid := msg.MessageID()

	mm = append(mm, overchanDelete(msgid))
	delmsgs := []string{}
	// get headers
	hdr, _ := self.daemon.database.GetHeadersForMessage(msgid)
	if hdr != nil {
		ref := hdr.Get("References", hdr.Get("Reference", ""))
		ref = strings.Trim(ref, "\t ")
		// is it a root post?
		if ref == "" {
			// load replies
			replies := self.daemon.database.GetThreadReplies(msgid, 0, 0)
			if replies != nil {
				for _, repl := range replies {
					// append mod line to mod message for reply
					mm = append(mm, overchanDelete(repl))
					// add to delete queue
					delmsgs = append(delmsgs, repl)
				}
			}
		}
	}
	delmsgs = append(delmsgs, msgid)
	// append mod line to mod message
	resp["deleted"] = delmsgs
	// only regen threads when we delete a non root port

	privkey_bytes := self.getSessionPrivkeyBytes(r)
	if privkey_bytes == nil {
		// crap this should never happen
		log.Println("failed to get private keys from session, not federating")
	} else {
		// wrap and sign mod message
		nntp, err := signArticle(wrapModMessage(mm), privkey_bytes)
		if err == nil {
			// send it off to federate
			self.modMessageChan <- nntp
		} else {
			resp["error"] = fmt.Sprintf("signing error: %s", err.Error())
		}
	}
	return resp
}

// ban the address of a poster
func (self httpModUI) HandleBanAddress(wr http.ResponseWriter, r *http.Request) {
	self.asAuthedWithMessage("ban", self.handleBanAddress, wr, r)
}

// delete a post
func (self httpModUI) HandleDeletePost(wr http.ResponseWriter, r *http.Request) {
	self.asAuthedWithMessage("login", self.handleDeletePost, wr, r)
}

func (self httpModUI) HandleLogin(wr http.ResponseWriter, r *http.Request) {
	privkey := r.FormValue("privkey")
	msg := "failed login: "
	if len(privkey) == 0 {
		msg += "no key"
	} else {
		ok, err := self.CheckKey(privkey, "login")
		if err != nil {
			msg += fmt.Sprintf("%s", err)
		} else if ok {
			msg = "login okay"
			sess := self.getSession(r)
			sess.Values["privkey"] = privkey
			sess.Save(r, wr)
		} else {
			msg += "invalid key"
		}
	}
	self.writeTemplateParam(wr, r, "modlogin_result.mustache", map[string]interface{}{"message": msg, csrf.TemplateTag: csrf.TemplateField(r)})
}

func (self httpModUI) HandleKeyGen(wr http.ResponseWriter, r *http.Request) {
	pk, sk := newSignKeypair()
	tripcode := makeTripcode(pk)
	self.writeTemplateParam(wr, r, "keygen.mustache", map[string]interface{}{"public": pk, "secret": sk, "tripcode": tripcode})
}

func (self httpModUI) ServeModPage(wr http.ResponseWriter, r *http.Request) {
	if self.checkSession(r, "login") {
		wr.Header().Set("X-CSRF-Token", csrf.Token(r))
		// we are logged in
		url := r.URL.String()
		if strings.HasSuffix(url, "/mod/feeds") {
			// serve feeds page
			self.writeTemplate(wr, r, "modfeed.mustache")
		} else {
			// serve mod page
			self.writeTemplate(wr, r, "modpage.mustache")
		}
	} else {
		// we are not logged in
		// serve login page
		self.writeTemplate(wr, r, "modlogin.mustache")
	}
	if r.Body != nil {
		r.Body.Close()
	}
}
