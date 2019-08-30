//
// frontend_http.go
//
// srnd http frontend implementation
//
package srnd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dchest/captcha"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"mime"
	"net/http"
	"strings"
	"time"
)

type bannedFunc func()
type errorFunc func(error)
type successFunc func(NNTPMessage)

// an attachment in a post
type postAttachment struct {
	DeleteFile func()
	NNTP       NNTPAttachment
	Filename   string `json:"name"`
	Filedata   string `json:"data"`
	Filetype   string `json:"type"`
}

// an api post request
type postRequest struct {
	Reference    string            `json:"reference"`
	Name         string            `json:"name"`
	Email        string            `json:"email"`
	Subject      string            `json:"subject"`
	Frontend     string            `json:"frontend"`
	Attachments  []postAttachment  `json:"files"`
	Group        string            `json:"newsgroup"`
	IpAddress    string            `json:"ip"`
	Destination  string            `json:"i2p"`
	Dubs         bool              `json:"dubs"`
	Message      string            `json:"message"`
	ExtraHeaders map[string]string `json:"headers"`
}

// regenerate a newsgroup page
type groupRegenRequest struct {
	// which newsgroup
	group string
	// page number
	page int
}

// livechan captcha solution
type liveCaptcha struct {
	ID       string
	Solution string
}

// inbound livechan command
type liveCommand struct {
	Type    string
	Captcha *liveCaptcha
	Post    *postRequest
}

type liveChan struct {
	// channel for recv-ing posts for sub'd newsgroup
	postchnl chan PostModel
	// channel for sending control messages
	datachnl chan []byte
	// unique session id
	uuid string
	// for recv-ing self
	resultchnl chan *liveChan
	// subbed newsgroup
	newsgroup string
	// have we solved captcha?
	captcha bool
	// our ip address
	IP string
}

// inform this livechan that we got a new post
func (lc *liveChan) Inform(post PostModel) {
	if lc.postchnl != nil {
		if lc.newsgroup == "" || lc.newsgroup == post.Board() {
			lc.postchnl <- post
		}
	}
}

func (lc *liveChan) SendError(err error) {
	msg, _ := json.Marshal(map[string]string{
		"Type":  "error",
		"Error": err.Error(),
	})
	if lc.datachnl != nil {
		lc.datachnl <- msg
	}
}

func (lc *liveChan) PostSuccess(nntp NNTPMessage) {
	// inform ui that a post was made
	msg, _ := json.Marshal(map[string]interface{}{
		"Type":  "posted",
		"Msgid": nntp.MessageID(),
		"OP":    nntp.OP(),
	})
	if lc.datachnl != nil {
		lc.datachnl <- msg
	}
}

func (lc *liveChan) SendBanned() {
	msg, _ := json.Marshal(map[string]string{
		"Type": "ban",
		// TODO: real ban message
		"Reason": "your an fagt, your IP was: " + lc.IP,
	})
	if lc.datachnl != nil {
		lc.datachnl <- msg
	}
}

// handle message from a websocket session
func (lc *liveChan) handleMessage(front *httpFrontend, cmd *liveCommand) {

	if cmd.Captcha != nil {
		lc.captcha = captcha.VerifyString(cmd.Captcha.ID, cmd.Captcha.Solution)
		// send captcha result
		msg, _ := json.Marshal(map[string]interface{}{
			"Type":    "captcha",
			"Success": lc.captcha,
		})
		if lc.datachnl != nil {
			lc.datachnl <- msg
		}
	}
	if lc.captcha && cmd.Post != nil {
		cmd.Post.Frontend = front.name
		cmd.Post.IpAddress = lc.IP
		if lc.newsgroup != "" {
			cmd.Post.Group = lc.newsgroup
		}
		cmd.Post.ExtraHeaders = map[string]string{"X-Livechan": "1"}
		front.handle_postRequest(cmd.Post, lc.SendBanned, lc.SendError, lc.PostSuccess, false)
	} else if cmd.Captcha == nil {
		// resend captcha challenge
		msg, _ := json.Marshal(map[string]string{
			"Type": "captcha",
		})

		if lc.datachnl != nil {
			lc.datachnl <- msg
		}

	}
}

type httpFrontend struct {
	modui    ModUI
	httpmux  *mux.Router
	daemon   *NNTPDaemon
	cache    CacheInterface
	bindaddr string
	name     string

	secret string

	webroot_dir  string
	template_dir string
	static_dir   string

	regen_threads  int
	regen_on_start bool
	attachments    bool

	prefix string

	store *sessions.CookieStore

	upgrader websocket.Upgrader

	jsonUsername        string
	jsonPassword        string
	enableJson          bool
	enableBoardCreation bool

	attachmentLimit int

	liveui_chnl       chan PostModel
	liveui_register   chan *liveChan
	liveui_deregister chan *liveChan
	end_liveui        chan bool
	// all liveui users
	// maps uuid -> liveChan
	liveui_chans     map[string]*liveChan
	liveui_usercount int

	// this is a very important thing by the way
	requireCaptcha bool

	// are we in archive mode?
	archive bool
}

// do we allow this newsgroup?
func (self httpFrontend) AllowNewsgroup(group string) bool {
	return newsgroupValidFormat(group) || group == "ctl" && !strings.HasSuffix(group, ".")
}

func (self *httpFrontend) Regen(msg ArticleEntry) {
	self.cache.Regen(msg)
}

func (self *httpFrontend) RegenerateBoard(board string) {
	self.cache.RegenerateBoard(board)
}

func (self *httpFrontend) RegenFrontPage() {
	pages, _ := self.daemon.database.GetUkkoPageCount(10)
	self.cache.RegenFrontPage(int(pages))
}

func (self httpFrontend) regenAll() {
	self.cache.RegenAll()
}

func (self httpFrontend) deleteThreadMarkup(root_post_id string) {
	self.cache.DeleteThreadMarkup(root_post_id)
}

func (self httpFrontend) deleteBoardMarkup(group string) {
	self.cache.DeleteBoardMarkup(group)
}

func (self *httpFrontend) ArchiveMode() {
	self.archive = true
}

// load post model and inform live ui
func (self *httpFrontend) informLiveUI(msgid, ref, group string) {
	// root post
	if ref == "" {
		ref = msgid
	}
	model := self.daemon.database.GetPostModel(self.prefix, msgid)
	// inform liveui
	if model != nil && self.liveui_chnl != nil {
		self.liveui_chnl <- model
		log.Println("liveui", msgid, ref, group)
	} else {
		log.Println("liveui failed to get model")
	}
}

// poll live ui events
func (self *httpFrontend) poll_liveui() {
	for {
		select {
		case live, ok := <-self.liveui_deregister:
			// deregister existing user event
			if ok {
				if self.liveui_chans != nil {
					delete(self.liveui_chans, live.uuid)
					self.liveui_usercount--
				}
				close(live.postchnl)
				live.postchnl = nil
				close(live.datachnl)
				live.datachnl = nil
			}
		case live, ok := <-self.liveui_register:
			// register new user event
			if ok {
				if self.liveui_chans != nil {
					live.uuid = randStr(10)
					live.postchnl = make(chan PostModel, 8)
					self.liveui_chans[live.uuid] = live
					self.liveui_usercount++
				}
				if live.resultchnl != nil {
					live.resultchnl <- live
					// get scrollback
					go func() {
						var threads []ThreadModel
						group := live.newsgroup
						if group == "" {
							// for ukko
							ents := self.daemon.database.GetLastBumpedThreads("", 5)
							if ents != nil {
								for _, e := range ents {
									g := e[1]
									page := self.daemon.database.GetGroupForPage(self.prefix, self.name, g, 0, 10)
									for _, t := range page.Threads() {
										if t.OP().MessageID() == e[0] {
											threads = append(threads, t)
											break
										}
									}
								}
							}
						} else {
							// for board
							board := self.daemon.database.GetGroupForPage(self.prefix, self.name, live.newsgroup, 0, 5)
							if board != nil {
								threads = board.Threads()
							}
						}

						if threads != nil {
							c := len(threads)
							for idx := range threads {
								th := threads[c-idx-1]
								th.Update(self.daemon.database)
								// send root post
								live.Inform(th.OP())
								// send replies
								for _, post := range th.Replies() {
									live.Inform(post)
								}
							}
						}
					}()
				}
			}
		case model, ok := <-self.liveui_chnl:

			// TODO: should we do board specific filtering?
			if ok {
				for _, livechan := range self.liveui_chans {
					go livechan.Inform(model)
				}
			}
		case <-self.end_liveui:
			livechnl := self.liveui_chnl
			self.liveui_chnl = nil
			close(livechnl)
			chnl := self.liveui_register
			self.liveui_register = nil
			close(chnl)
			chnl = self.liveui_deregister
			self.liveui_deregister = nil
			close(chnl)
			// remove all
			for _, livechan := range self.liveui_chans {
				if livechan.postchnl != nil {
					close(livechan.postchnl)
					livechan.postchnl = nil
				}
			}
			self.liveui_chans = nil
			return
		}
	}
}

func (self *httpFrontend) poll() {

	// regenerate front page
	self.RegenFrontPage()

	// trigger regen
	if self.regen_on_start {
		self.cache.RegenAll()
	}

	modChnl := self.modui.MessageChan()
	for {
		select {
		case nntp := <-modChnl:
			storeMessage(self.daemon, nntp.MIMEHeader(), nntp.BodyReader())
		}
	}
}

func (self *httpFrontend) HandleNewPost(nntp frontendPost) {
	msgid := nntp.MessageID()
	group := nntp.Newsgroup()
	ref := nntp.Reference()
	go self.informLiveUI(msgid, ref, group)
	if len(ref) > 0 {
		msgid = ref
	}

	entry := ArticleEntry{msgid, group}
	// regnerate thread
	self.Regen(entry)
	// regenerate all board pages if not archiving
	if !self.archive {
		self.RegenerateBoard(group)
	}
	// regen front page
	self.RegenFrontPage()

}

// create a new captcha, return as json object
func (self *httpFrontend) new_captcha_json(wr http.ResponseWriter, r *http.Request) {
	s, err := self.store.Get(r, self.name)
	if err != nil {
		http.Error(wr, err.Error(), 500)
		return
	}
	captcha_id := captcha.New()
	resp := make(map[string]string)
	// the captcha id
	resp["id"] = captcha_id
	s.Values["captcha_id"] = captcha_id
	s.Save(r, wr)
	// url of the image
	resp["url"] = fmt.Sprintf("%scaptcha/%s.png", self.prefix, captcha_id)
	wr.Header().Set("Content-Type", "text/json; encoding=UTF-8")
	enc := json.NewEncoder(wr)
	enc.Encode(&resp)
}

// handle newboard page
func (self *httpFrontend) handle_newboard(wr http.ResponseWriter, r *http.Request) {
	param := make(map[string]interface{})
	param["prefix"] = self.prefix
	io.WriteString(wr, template.renderTemplate("newboard", param, self.cache.GetHandler().GetI18N(r)))
}

// handle new post via http request for a board
func (self *httpFrontend) handle_postform(wr http.ResponseWriter, r *http.Request, board string, sendJson, checkCaptcha bool) {

	// the post we will turn into an nntp article
	pr := new(postRequest)
	pr.ExtraHeaders = make(map[string]string)

	if sendJson {
		wr.Header().Add("Content-Type", "text/json; encoding=UTF-8")
	}

	// close request body when done
	defer r.Body.Close()

	mp_reader, err := r.MultipartReader()

	if err != nil {
		wr.WriteHeader(500)
		if sendJson {
			json.NewEncoder(wr).Encode(map[string]interface{}{"error": err.Error()})
		} else {
			io.WriteString(wr, err.Error())
		}
		return
	}

	pr.Group = board

	// encrypt IP Addresses
	// when a post is recv'd from a frontend, the remote address is given its own symetric key that the local srnd uses to encrypt the address with, for privacy
	// when a mod event is fired, it includes the encrypted IP address and the symetric key that frontend used to encrypt it, thus allowing others to determine the IP address
	// each node will optionally comply with the mod event, banning the address from being able to post from that frontend

	// get the "real" ip address from the request
	pr.IpAddress, err = extractRealIP(r)
	pr.Destination = r.Header.Get("X-I2P-DestHash")
	pr.Frontend = self.name

	var captcha_retry bool
	var captcha_solution, captcha_id string
	url := self.generateBoardURL(board, 0)
	var part_buff bytes.Buffer
	for {
		part, err := mp_reader.NextPart()
		if err == nil {
			defer part.Close()
			// get the name of the part
			partname := part.FormName()
			// read part for attachment
			if strings.HasPrefix(partname, "attachment_") && self.attachments {
				if len(pr.Attachments) < self.attachmentLimit {
					att := readAttachmentFromMimePartAndStore(part, self.daemon.store)
					if att != nil && att.Filename() != "" {
						log.Println("attaching file", att.Filename())
						pa := postAttachment{
							Filename: att.Filename(),
							Filetype: att.Mime(),
							NNTP:     att,
							DeleteFile: func() {
								f := att.Filepath()
								DelFile(self.daemon.store.AttachmentFilepath(f))
								DelFile(self.daemon.store.ThumbnailFilepath(f))
							},
						}
						pr.Attachments = append(pr.Attachments, pa)
					}
				}
				continue
			}
			io.Copy(&part_buff, part)

			// check for values we want
			if partname == "subject" {
				pr.Subject = part_buff.String()
			} else if partname == "name" {
				pr.Name = part_buff.String()
			} else if partname == "message" {
				pr.Message = strings.Trim(part_buff.String(), "\r")
			} else if partname == "reference" {
				pr.Reference = part_buff.String()
				if len(pr.Reference) != 0 {
					url = self.generateThreadURL(pr.Reference)
				}
			} else if partname == "captcha_id" {
				captcha_id = part_buff.String()
			} else if partname == "captcha" {
				captcha_solution = part_buff.String()
			} else if partname == "dubs" {
				pr.Dubs = part_buff.String() == "on"
			} else if partname == "uri" {
				str := part_buff.String()
				if len(str) > 0 {
					pr.ExtraHeaders["X-References-Uri"] = safeHeader(str)
				}
			}

			// we done
			// reset buffer for reading parts
			part_buff.Reset()
		} else {
			if err != io.EOF {
				for _, att := range pr.Attachments {
					if att.DeleteFile != nil {
						att.DeleteFile()
					}
				}
				// TODO: we need to delete uploaded files somehow since they are unregistered at this point
				errmsg := fmt.Sprintf("httpfrontend post handler error reading multipart: %s", err)
				log.Println(errmsg)
				wr.WriteHeader(500)
				if sendJson {
					json.NewEncoder(wr).Encode(map[string]interface{}{"error": errmsg})
				} else {
					io.WriteString(wr, errmsg)
				}
				return
			}
			break
		}
	}

	sess, err := self.store.Get(r, self.name)
	if err != nil {
		errmsg := fmt.Sprintf("session store error: %s", err.Error())
		if sendJson {
			json.NewEncoder(wr).Encode(map[string]interface{}{"error": errmsg})
		} else {
			io.WriteString(wr, errmsg)
		}
		return
	}
	if checkCaptcha && len(captcha_id) == 0 {
		cid, ok := sess.Values["captcha_id"]
		if ok {
			captcha_id = cid.(string)
		} else {
			log.Println("no captcha id in session?")
		}
		sess.Values["captcha_id"] = ""
	}
	log.Println("captcha", captcha_id, "try '", captcha_solution, "'")
	if checkCaptcha && !captcha.VerifyString(captcha_id, captcha_solution) {
		// captcha is not valid
		captcha_retry = true
	} else {
		// valid captcha
		// increment post count
		var posts int
		val, ok := sess.Values["posts"]
		if ok {
			posts = val.(int)
		} else {
			posts = 0
		}
		posts++
		sess.Values["posts"] = posts
	}
	sess.Save(r, wr)

	// make error template param
	resp_map := make(map[string]interface{})
	resp_map["prefix"] = self.prefix
	// set redirect url
	if len(url) > 0 {
		// if we explicitly know the url use that
		resp_map["redirect_url"] = url
	} else {
		// if our referer is saying we are from /new/ page use that
		// otherwise use prefix
		if strings.HasSuffix(r.Referer(), self.prefix+"new/") {
			resp_map["redirect_url"] = self.prefix + "new/"
		} else {
			resp_map["redirect_url"] = self.prefix
		}
	}

	if captcha_retry {
		for _, att := range pr.Attachments {
			if att.DeleteFile != nil {
				att.DeleteFile()
			}
		}
		if sendJson {
			json.NewEncoder(wr).Encode(map[string]interface{}{"error": "bad captcha"})
		} else {
			// retry the post with a new captcha
			resp_map = make(map[string]interface{})
			resp_map["prefix"] = self.prefix
			resp_map["redirect_url"] = url
			resp_map["reason"] = "captcha incorrect"
			io.WriteString(wr, template.renderTemplate("post_fail", resp_map, self.cache.GetHandler().GetI18N(r)))
		}
		return
	}

	b := func() {
		for _, att := range pr.Attachments {
			if att.DeleteFile != nil {
				att.DeleteFile()
			}
		}
		if sendJson {
			wr.WriteHeader(403)
			json.NewEncoder(wr).Encode(map[string]interface{}{"error": "banned"})
		} else {
			wr.WriteHeader(403)
			io.WriteString(wr, "banned")
		}
	}

	e := func(err error) {
		for _, att := range pr.Attachments {
			if att.DeleteFile != nil {
				att.DeleteFile()
			}
		}
		log.Println("frontend error:", err)
		wr.WriteHeader(200)
		if sendJson {
			json.NewEncoder(wr).Encode(map[string]interface{}{"error": err.Error()})
		} else {
			resp_map["reason"] = err.Error()
			resp_map["prefix"] = self.prefix
			resp_map["redirect_url"] = url
			io.WriteString(wr, template.renderTemplate("post_fail", resp_map, self.cache.GetHandler().GetI18N(r)))
		}
	}

	s := func(nntp NNTPMessage) {
		// send success reply
		wr.WriteHeader(201)
		// determine the root post so we can redirect to the thread for it
		msg_id := nntp.Headers().Get("References", nntp.MessageID())
		// render response as success
		url := self.generateThreadURL(msg_id)
		if sendJson {
			json.NewEncoder(wr).Encode(map[string]interface{}{"message_id": nntp.MessageID(), "url": url, "error": nil})
		} else {
			template.writeTemplate("post_success", map[string]interface{}{"prefix": self.prefix, "message_id": nntp.MessageID(), "redirect_url": url}, wr, self.cache.GetHandler().GetI18N(r))
		}
	}
	self.handle_postRequest(pr, b, e, s, self.enableBoardCreation)
}

func (self *httpFrontend) generateThreadURL(msgid string) (url string) {
	url = fmt.Sprintf("%st/%s/", self.prefix, HashMessageID(msgid))
	return
}

func (self *httpFrontend) generateBoardURL(newsgroup string, pageno int) (url string) {
	if pageno > 0 {
		url = fmt.Sprintf("%sb/%s/%d/", self.prefix, newsgroup, pageno)
	} else {
		url = fmt.Sprintf("%sb/%s/", self.prefix, newsgroup)
	}
	return
}

// turn a post request into an nntp article write it to temp dir and tell daemon
func (self *httpFrontend) handle_postRequest(pr *postRequest, b bannedFunc, e errorFunc, s successFunc, createGroup bool) {
	var err error
	if len(pr.Attachments) > self.attachmentLimit {
		err = errors.New("too many attachments")
		e(err)
		return
	}
	pr.Message = strings.Trim(pr.Message, "\r")
	m := strings.Trim(pr.Message, "\n")
	m = strings.Trim(m, " ")
	if len(pr.Attachments) == 0 && len(m) == 0 {
		err = errors.New("no post message")
		e(err)
		return
	}
	nntp := new(nntpArticle)
	defer nntp.Reset()
	var banned bool
	nntp.headers = make(ArticleHeaders)
	address := pr.IpAddress
	// check for banned
	if len(address) > 0 {
		banned, err = self.daemon.database.CheckIPBanned(address)
		if err == nil {
			if banned {
				b()
				return
			}
		} else {
			e(err)
			return
		}
	}
	if len(address) == 0 {
		address = "Tor"
	}

	if !strings.HasPrefix(address, "127.") {
		// set the ip address of the poster to be put into article headers
		// if we cannot determine it, i.e. we are on Tor/i2p, this value is not set
		if address == "Tor" {
			nntp.headers.Set("X-Tor-Poster", "1")
		} else {
			address, err = self.daemon.database.GetEncAddress(address)
			if err == nil {
				nntp.headers.Set("X-Encrypted-IP", address)
			} else {
				e(err)
				return
			}
			// TODO: add x-tor-poster header for tor exits
		}
	}

	board := pr.Group

	// post fail message
	banned, err = self.daemon.database.NewsgroupBanned(board)
	if banned {
		e(errors.New("newsgroup banned "))
		return
	}
	if err != nil {
		e(err)
	}

	if !createGroup && !self.daemon.database.HasNewsgroup(board) {
		e(errors.New("we don't have this newsgroup " + board))
		return
	}

	// if we don't have an address for the poster try checking for i2p httpd headers
	if len(pr.Destination) == i2pDestHashLen() {
		nntp.headers.Set("X-I2P-DestHash", pr.Destination)
	}

	ref := pr.Reference
	if ref != "" {
		if ValidMessageID(ref) {
			if self.daemon.database.HasArticleLocal(ref) {
				nntp.headers.Set("References", ref)
			} else {
				e(errors.New("article referenced not locally available"))
				return
			}
		} else {
			e(errors.New("invalid reference"))
			return
		}
	}

	// set newsgroup
	nntp.headers.Set("Newsgroups", pr.Group)

	// check message size
	if len(pr.Attachments) == 0 && len(pr.Message) == 0 {
		e(errors.New("no message"))
		return
	}
	// TODO: make configurable
	if len(pr.Message) > 1024*1024 {
		e(errors.New("your message is too big"))
		return
	}

	if !self.daemon.CheckText(pr.Message) {
		e(errors.New("spam"))
		return
	}

	if len(pr.Frontend) == 0 {
		// :-DDD
		pr.Frontend = "mongo.db.is.web.scale"
	} else if len(pr.Frontend) > 128 {
		e(errors.New("frontend name is too long"))
		return
	}

	subject := strings.TrimSpace(pr.Subject)

	// set subject
	if subject == "" {
		subject = "None"
	} else if len(subject) > 256 {
		// subject too big
		e(errors.New("Subject is too long"))
		return
	}

	nntp.headers.Set("Subject", safeHeader(subject))
	if isSage(subject) && ref != "" {
		nntp.headers.Set("X-Sage", "1")
	}

	name := strings.TrimSpace(pr.Name)
	var tripcode_privkey []byte
	// tripcode
	if idx := strings.IndexByte(name, '#'); idx >= 0 {
		tripcode_privkey = parseTripcodeSecret(name[idx+1:])
		name = strings.TrimSpace(name[:idx])
	}
	if name == "" {
		name = "Anonymous"
	}
	if len(name) > 128 {
		// name too long
		e(errors.New("name too long"))
		return
	}
	msgid := genMessageID(pr.Frontend)
	// roll until dubs if desired
	for pr.Dubs && !MessageIDWillDoDubs(msgid) {
		msgid = genMessageID(pr.Frontend)
	}

	nntp.headers.Set("From", formatAddress(safeHeader(name), "poster@"+pr.Frontend))
	nntp.headers.Set("Message-ID", msgid)

	// set message
	nntp.message = nntpSanitize(pr.Message)

	cites, err := self.daemon.database.FindCitesInText(pr.Message)
	if err != nil {
		e(err)
		return
	}

	if len(cites) > 0 {
		if ref == "" && len(cites) == 1 {
			/*
				this is workaround for:

				{RFC 5322}
				If the parent message does not contain
				a "References:" field but does have an "In-Reply-To:" field
				containing a single message identifier, then the "References:" field
				will contain the contents of the parent's "In-Reply-To:" field
				followed by the contents of the parent's "Message-ID:" field (if
				any).
			*/
			cites = append(cites, "<0>")
		}
		nntp.headers.Set("In-Reply-To", strings.Join(cites, " "))
	}

	// set date
	nntp.headers.Set("Date", timeNowStr())
	// append path from frontend
	nntp.AppendPath(pr.Frontend)

	// add extra headers if needed
	if pr.ExtraHeaders != nil {
		for name, val := range pr.ExtraHeaders {
			// don't overwrite existing headers
			if nntp.headers.Get(name, "") == "" {
				nntp.headers.Set(name, val)
			}
		}
	}
	if self.attachments {
		var delfiles []string
		for _, att := range pr.Attachments {
			// add attachment
			if len(att.Filedata) > 0 {
				a := createAttachment(att.Filetype, att.Filetype, strings.NewReader(att.Filedata))
				nntp.Attach(a)
				err = a.Save(self.daemon.store.AttachmentDir())
				if err == nil {
					delfiles = append(delfiles, a.Filepath())
					// check if we need to thumbnail it
					if !CheckFile(self.daemon.store.ThumbnailFilepath(a.Filepath())) {
						_, err = self.daemon.store.GenerateThumbnail(a.Filepath())
					}
					if err == nil {
						delfiles = append(delfiles, self.daemon.store.ThumbnailFilepath(a.Filepath()))
					}
				}
				if err != nil {
					break
				}
			} else {
				nntp.Attach(att.NNTP)
			}
		}
		if err != nil {
			// nuke files
			for _, fname := range delfiles {
				DelFile(fname)
			}
			for _, att := range pr.Attachments {
				if att.DeleteFile != nil {
					att.DeleteFile()
				}
			}
			e(err)
			return
		}
	}
	// pack it before sending so that the article is well formed
	// sign if needed
	if len(tripcode_privkey) == 32 {
		pk, _ := naclSeedToKeyPair(tripcode_privkey)
		nntp.headers.Set("X-PubKey-Ed25519", hexify(pk))
		nntp.Pack()
		if err != nil {
			e(err)
			return
		}
		nntp, err = signArticle(nntp, tripcode_privkey)
		if err != nil {
			// error signing
			e(err)
			return
		}
	} else {
		nntp.Pack()
	}
	// have daemon sign message-id
	self.daemon.WrapSign(nntp)

	err = storeMessage(self.daemon, nntp.MIMEHeader(), nntp.BodyReader())

	if err != nil {
		// clean up
		self.daemon.expire.ExpirePost(nntp.MessageID())
		e(err)
	} else {
		s(nntp)
	}
}

// handle posting / postform
func (self httpFrontend) handle_poster(wr http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	sendJSON := false
	if strings.HasSuffix(path, "/json") || strings.HasSuffix(path, "/json/") {
		sendJSON = true
	}
	var board string
	// extract board
	parts := strings.Count(path, "/")
	if parts > 1 {
		board = strings.Split(path, "/")[2]
	}

	// this is a POST request
	if r.Method == "POST" && self.AllowNewsgroup(board) && newsgroupValidFormat(board) {
		self.handle_postform(wr, r, board, sendJSON, self.requireCaptcha)
	} else {
		wr.WriteHeader(403)
		io.WriteString(wr, "Nope")
	}
}

func (self *httpFrontend) serve_captcha(wr http.ResponseWriter, r *http.Request) {
	s, err := self.store.Get(r, self.name)
	if err == nil {
		captcha_id := captcha.New()
		s.Values["captcha_id"] = captcha_id
		log.Println("captcha_id", captcha_id)
		s.Save(r, wr)
		redirect_url := fmt.Sprintf("%scaptcha/%s.png", self.prefix, captcha_id)
		// redirect to the image
		http.Redirect(wr, r, redirect_url, 302)
	} else {
		// handle session error
		// TODO: clear cookies?
		http.Error(wr, err.Error(), 500)
	}
}

// send error
func api_error(wr http.ResponseWriter, err error) {
	resp := make(map[string]string)
	resp["error"] = err.Error()
	wr.Header().Add("Content-Type", "text/json; encoding=UTF-8")
	enc := json.NewEncoder(wr)
	enc.Encode(resp)
}

// authenticated part of api
// handle all functions that require authentication
func (self httpFrontend) handle_authed_api(wr http.ResponseWriter, r *http.Request, api string) {
	// check valid format
	ct := strings.ToLower(r.Header.Get("Content-Type"))
	mtype, _, err := mime.ParseMediaType(ct)
	if err == nil {
		if strings.HasSuffix(mtype, "/json") {
			// valid :^)
		} else {
			// bad content type
			api_error(wr, errors.New(fmt.Sprintf("invalid content type: %s", ct)))
			return
		}
	} else {
		// bad content type
		api_error(wr, err)
		return
	}

	b := func() {
		api_error(wr, errors.New("banned"))
	}

	e := func(err error) {
		api_error(wr, err)
	}

	s := func(nntp NNTPMessage) {
		wr.Header().Add("Content-Type", "text/json; encoding=UTF-8")
		resp := make(map[string]string)
		resp["id"] = nntp.MessageID()
		enc := json.NewEncoder(wr)
		enc.Encode(resp)
	}

	dec := json.NewDecoder(r.Body)
	if api == "post" {
		var pr postRequest
		err = dec.Decode(&pr)
		r.Body.Close()
		if err == nil {
			// we parsed it
			self.handle_postRequest(&pr, b, e, s, true)
		} else {
			// bad parsing?
			api_error(wr, err)
		}
	} else {
		// no such method
		wr.WriteHeader(404)
		io.WriteString(wr, "No such method")
	}
}

// handle find post api command
func (self *httpFrontend) handle_api_find(wr http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	h := q.Get("hash")
	msgid := q.Get("id")
	if len(h) > 0 {
		e, err := self.daemon.database.GetMessageIDByHash(h)
		if err == nil {
			msgid = e.MessageID()
		}
	}

	if !ValidMessageID(msgid) {
		msgid = ""
	}

	if len(msgid) > 0 {
		self.daemon.store.GetMessage(msgid, func(nntp NNTPMessage) {
			if nntp == nil {
				wr.WriteHeader(404)
				return
			}
			model := PostModelFromMessage(self.prefix, nntp)
			// we found it
			wr.Header().Add("Content-Type", "text/json; encoding=UTF-8")
			json.NewEncoder(wr).Encode([]PostModel{model})
		})
		return
	}
	s := q.Get("text")
	g := q.Get("group")

	wr.Header().Add("Content-Type", "text/json; encoding=UTF-8")
	chnl := make(chan PostModel)
	wr.WriteHeader(http.StatusOK)
	io.WriteString(wr, "[")
	donechnl := make(chan int)
	go func(w io.Writer) {
		for {
			p, ok := <-chnl
			if ok {
				d, e := json.Marshal(p)
				if e == nil {
					io.WriteString(w, string(d))
					io.WriteString(w, ", ")
				} else {
					log.Println("error marshalling post", e)
				}
			} else {
				break
			}
		}
		donechnl <- 0
	}(wr)
	limit := 50
	if len(h) > 0 {
		go self.daemon.database.SearchByHash(self.prefix, g, h, chnl, limit)
	} else {
		go self.daemon.database.SearchQuery(self.prefix, g, s, chnl, limit)
	}
	chnl <- nil
	<-donechnl
	io.WriteString(wr, " null ]")
	return
}

// handle un authenticated part of api
func (self *httpFrontend) handle_unauthed_api(wr http.ResponseWriter, r *http.Request, api string) {
	var err error
	if api == "header" {
		var msgids []string
		q := r.URL.Query()
		name := strings.ToLower(q.Get("name"))
		val := q.Get("value")
		msgids, err = self.daemon.database.GetMessageIDByHeader(name, val)
		if err == nil {
			wr.Header().Add("Content-Type", "text/json; encoding=UTF-8")
			json.NewEncoder(wr).Encode(msgids)
		} else {
			api_error(wr, err)
		}
	} else if api == "groups" {
		wr.Header().Add("Content-Type", "text/json; encoding=UTF-8")
		groups := self.daemon.database.GetAllNewsgroups()
		json.NewEncoder(wr).Encode(groups)
	} else if api == "find" {
		self.handle_api_find(wr, r)
	} else if api == "history" {
		var s PostingStats
		q := r.URL.Query()
		now := timeNow()
		begin := queryGetInt64(q, "start", now-(48*3600000))
		end := queryGetInt64(q, "end", now)
		gran := queryGetInt64(q, "granularity", 3600000)
		s, err = self.daemon.database.GetPostingStats(gran, begin, end)
		if err == nil {
			wr.Header().Add("Content-Type", "text/json; encoding=UTF-8")
			json.NewEncoder(wr).Encode(s)
		} else {
			api_error(wr, err)
		}
	}
}

// handle livechan api
func (self *httpFrontend) handle_liveapi(w http.ResponseWriter, r *http.Request) {
	// response
	res := make(map[string]interface{})
	// set content type
	w.Header().Set("Content-Type", "text/json; encoding=UTF-8")

	// check for ip banned
	ip, err := extractRealIP(r)
	if err == nil {
		var banned bool
		banned, err = self.daemon.database.CheckIPBanned(ip)
		if banned {
			// TODO: ban reason
			res["error"] = "u banned yo"
			w.WriteHeader(http.StatusForbidden)
		} else if err == nil {
			// obtain session
			s, err := self.store.Get(r, self.name)
			if err == nil {
				vars := mux.Vars(r)
				meth := vars["meth"]
				if r.Method == "POST" {
					if meth == "captcha" {
						// /livechan/api/captcha

						// post captcha solution
						c := new(liveCaptcha)
						defer r.Body.Close()
						dec := json.NewDecoder(r.Body)
						err = dec.Decode(c)
						if err == nil {
							// decode success
							res["success"] = false
							if captcha.VerifyString(c.ID, c.Solution) {
								// successful captcha
								res["success"] = true
								s.Values["captcha"] = true
							}
						} else {
							// decode error
							res["error"] = err.Error()
							s.Values["captcha"] = false
						}
						s.Save(r, w)
					} else if meth == "post" {
						// /livechan/api/post?newsgroup=overchan.boardname

						// post to a board
						board := r.URL.Query().Get("newsgroup")
						if self.AllowNewsgroup(board) && newsgroupValidFormat(board) {

							// check if we solved captcha
							val, ok := s.Values["captcha"]
							if ok {
								var live bool
								if live, ok = val.(bool); ok && live {
									// treat as frontend post
									// send json and bypass checking for captcha in request body
									self.handle_postform(w, r, board, true, false)
									// done
									return
								} else {
									// not livechan or captcha is not filled out
									res["captcha"] = true
								}

							} else {
								// not a livechan session
								res["captcha"] = true
							}

						} else {
							// bad newsgroup
							res["error"] = "bad newsgroup: " + board
						}
					} else {
						// bad post method
						res["error"] = "no such method: " + meth
					}
				} else if r.Method == "GET" {
					// handle GET methods for api endpoint
					if meth == "online" {
						// /livechan/api/online

						// return how many users are online
						res["online"] = self.liveui_usercount
					} else if meth == "pph" {
						// /livechan/api/pph?newsgroup=overchan.boardname

						// return post per hour count
						// TODO: implement better (?)
						board := r.URL.Query().Get("newsgroup")
						if newsgroupValidFormat(board) {
							res["pph"] = self.daemon.database.CountPostsInGroup(board, 3600)
						} else {
							res["error"] = "invalid newsgroup"
						}
					} else {
						// unknown method
						res["error"] = "unknown method: " + meth
					}
				} else {
					// bad method ( should never happen tho, catch case regardless )
					res["error"] = "not found"
					w.WriteHeader(http.StatusMethodNotAllowed)
				}
			} else {
				// failed to get session
				res["error"] = "could not parse session: " + err.Error()
				w.WriteHeader(http.StatusBadRequest)
			}
		} else {
			// ban error check
			res["error"] = "error checking ban: " + err.Error()
			w.WriteHeader(http.StatusInternalServerError)
		}
	} else {
		// could not extract ip address
		res["error"] = "could not extract ip: " + err.Error()
		w.WriteHeader(http.StatusBadRequest)
	}
	// write response
	enc := json.NewEncoder(w)
	enc.Encode(res)
}

func (self *httpFrontend) handle_api(wr http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	meth := vars["meth"]
	if r.Method == "POST" && self.enableJson {
		u, p, ok := r.BasicAuth()
		if ok && u == self.jsonUsername && p == self.jsonPassword {
			// authenticated
			self.handle_authed_api(wr, r, meth)
		} else {
			// invalid auth
			wr.WriteHeader(401)
		}
	} else if r.Method == "GET" {
		self.handle_unauthed_api(wr, r, meth)
	} else {
		wr.WriteHeader(404)
	}
}

// upgrade to web sockets and subscribe to all new posts
func (self *httpFrontend) handle_liveui(w http.ResponseWriter, r *http.Request) {

	IpAddress, err := extractRealIP(r)
	log.Println("liveui:", IpAddress)
	var banned bool
	if err == nil {
		banned, err = self.daemon.database.CheckIPBanned(IpAddress)
		if banned {
			w.WriteHeader(403)
			io.WriteString(w, "banned")
			return
		}
	}

	if err != nil {
		w.WriteHeader(504)
		log.Println("parse ip:", err)
		io.WriteString(w, err.Error())
		return
	}

	conn, err := self.upgrader.Upgrade(w, r, nil)
	if err != nil {
		// problem
		w.WriteHeader(500)
		io.WriteString(w, err.Error())
		return
	}
	// obtain a new channel for reading post models
	board := ""
	if r.URL.RawQuery != "" {
		board = "overchan." + r.URL.RawQuery
	}
	livechnl := self.subscribe(board, IpAddress)
	if livechnl == nil {
		// shutting down
		conn.Close()
		return
	}
	// okay we got a channel
	live := <-livechnl
	close(livechnl)
	go func() {
		// read loop
		for {
			_, _, err := conn.NextReader()
			if err != nil {
				conn.Close()
				if self.liveui_deregister != nil {
					self.liveui_deregister <- live
				}
				return
			}
		}
	}()
	ticker := time.NewTicker(time.Second * 5)
	for err == nil {
		select {
		case model, ok := <-live.postchnl:
			if ok && model != nil {
				err = conn.WriteJSON(model)
			} else {
				// channel closed
				break
			}
		case data, ok := <-live.datachnl:
			if ok {
				err = conn.WriteMessage(websocket.TextMessage, data)
			} else {
				break
			}
		case <-ticker.C:
			conn.WriteMessage(websocket.PingMessage, []byte{})
		}
	}
	conn.Close()
}

// get a chan that is subscribed to all new posts in a newsgroup
func (self *httpFrontend) subscribe(board, ip string) chan *liveChan {
	if self.liveui_register == nil {
		return nil
	} else {
		live := new(liveChan)
		live.IP = ip
		live.newsgroup = board
		live.resultchnl = make(chan *liveChan)
		live.datachnl = make(chan []byte, 8)
		self.liveui_register <- live
		return live.resultchnl
	}
}

func (self *httpFrontend) Mainloop() {
	EnsureDir(self.webroot_dir)
	if !CheckFile(self.template_dir) {
		log.Fatalf("no such template folder %s", self.template_dir)
	}
	template.changeTemplateDir(self.template_dir)

	// set up handler mux
	self.httpmux = mux.NewRouter()

	self.httpmux.NotFoundHandler = template.createNotFoundHandler(self.prefix, self.name)

	// create mod ui
	self.modui = createHttpModUI(self)

	cache_handler := self.cache.GetHandler()

	// csrf protection
	b := []byte(self.secret)
	var sec [32]byte
	copy(sec[:], b)
	// TODO: make configurable
	CSRF := csrf.Protect(sec[:], csrf.Secure(false))

	m := mux.NewRouter()
	// modui handlers
	m.Path("/mod/").HandlerFunc(self.modui.ServeModPage).Methods("GET")
	m.Path("/mod/feeds").HandlerFunc(self.modui.ServeModPage).Methods("GET")
	m.Path("/mod/keygen").HandlerFunc(self.modui.HandleKeyGen).Methods("GET")
	m.Path("/mod/login").HandlerFunc(self.modui.HandleLogin).Methods("POST")
	m.Path("/mod/spam").HandlerFunc(self.modui.HandlePostSpam).Methods("POST")
	m.Path("/mod/del/{article_hash}").HandlerFunc(self.modui.HandleDeletePost).Methods("GET")
	m.Path("/mod/ban/{address}").HandlerFunc(self.modui.HandleBanAddress).Methods("GET")
	m.Path("/mod/unban/{address}").HandlerFunc(self.modui.HandleUnbanAddress).Methods("GET")
	m.Path("/mod/addkey/{pubkey}").HandlerFunc(self.modui.HandleAddPubkey).Methods("GET")
	m.Path("/mod/delkey/{pubkey}").HandlerFunc(self.modui.HandleDelPubkey).Methods("GET")
	m.Path("/mod/admin/{action}").HandlerFunc(self.modui.HandleAdminCommand).Methods("GET", "POST")
	self.httpmux.PathPrefix("/mod/").Handler(CSRF(m))
	m = self.httpmux
	// robots.txt handler
	m.Path("/robots.txt").HandlerFunc(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		io.WriteString(w, "User-Agent: *\nDisallow: /\n")
	})).Methods("GET")

	m.Path("/thm/{f}").Handler(http.FileServer(http.Dir(self.webroot_dir)))
	m.Path("/img/{f}").Handler(http.FileServer(http.Dir(self.webroot_dir)))
	m.PathPrefix("/b/").Handler(cache_handler).Methods("GET", "HEAD")
	m.PathPrefix("/t/").Handler(cache_handler).Methods("GET", "HEAD")
	m.Path("/{f}.html").Handler(cache_handler).Methods("GET", "HEAD")
	m.Path("/{f}.json").Handler(cache_handler).Methods("GET", "HEAD")
	m.PathPrefix("/o/").Handler(cache_handler).Methods("GET", "HEAD")
	m.PathPrefix("/overboard/").Handler(cache_handler).Methods("GET", "HEAD")
	m.PathPrefix("/static/").Handler(http.FileServer(http.Dir(self.static_dir)))
	m.PathPrefix("/post/").HandlerFunc(self.handle_poster).Methods("POST")
	m.Path("/captcha/new").HandlerFunc(self.new_captcha_json).Methods("GET")
	m.Path("/captcha/img").HandlerFunc(self.serve_captcha).Methods("GET")
	m.Path("/captcha/{f}").Handler(captcha.Server(350, 175)).Methods("GET")
	m.Path("/new/").HandlerFunc(self.handle_newboard).Methods("GET")
	m.Path("/api/{meth}").HandlerFunc(self.handle_api).Methods("POST", "GET")
	// live ui websocket
	m.Path("/live").HandlerFunc(self.handle_liveui).Methods("GET")
	// live ui page
	m.Path("/livechan/").HandlerFunc(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		template.writeTemplate("live", map[string]interface{}{"prefix": self.prefix}, w, self.cache.GetHandler().GetI18N(r))
	})).Methods("GET", "HEAD")
	// live ui api endpoint
	m.Path("/livechan/api/{meth}").HandlerFunc(self.handle_liveapi).Methods("GET", "POST")

	m.Path("/").Handler(cache_handler)

	var err error

	// run daemon's mod engine with our frontend
	// go RunModEngine(self.daemon.mod, self.cache.RegenOnModEvent)

	if self.archive {
		self.cache.InvertPagination()
	}

	// start cache
	self.cache.Start()

	// this is for link cache
	// go template.loadAllModels(self.prefix, self.name, self.daemon.database)

	// poll channels

	go self.poll()

	// poll liveui
	go self.poll_liveui()

	// start webserver here
	log.Printf("frontend %s binding to %s", self.name, self.bindaddr)

	// serve it!
	err = http.ListenAndServe(self.bindaddr, self.httpmux)
	if err != nil {
		log.Fatalf("failed to bind frontend %s %s", self.name, err)
	}
}

func (self *httpFrontend) endLiveUI() {
	// end live ui
	if self.end_liveui != nil {
		self.end_liveui <- true
		close(self.end_liveui)
		self.end_liveui = nil
	}
}

func (self *httpFrontend) RegenOnModEvent(newsgroup, msgid, root string, page int) {
	self.cache.RegenOnModEvent(newsgroup, msgid, root, page)
}

func (self *httpFrontend) GetCacheHandler() CacheHandler {
	return self.cache.GetHandler()
}

// create a new http based frontend
func NewHTTPFrontend(daemon *NNTPDaemon, cache CacheInterface, config map[string]string, url string) Frontend {
	template.Minimize = config["minimize_html"] == "1"
	front := new(httpFrontend)
	front.daemon = daemon
	front.cache = cache
	front.attachments = mapGetInt(config, "allow_files", 1) == 1
	front.bindaddr = config["bind"]
	front.name = config["name"]
	front.webroot_dir = config["webroot"]
	front.static_dir = config["static_files"]
	front.template_dir = config["templates"]
	front.prefix = config["prefix"]
	front.regen_on_start = config["regen_on_start"] == "1"
	front.enableBoardCreation = config["board_creation"] == "1"
	front.requireCaptcha = config["rapeme"] != "omgyesplz"
	cache.SetRequireCaptcha(front.requireCaptcha)
	if config["json-api"] == "1" {
		front.jsonUsername = config["json-api-username"]
		front.jsonPassword = config["json-api-password"]
		front.enableJson = true
	}
	front.upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			// TODO: detect origin
			return true
		},
	}
	front.attachmentLimit = 5
	front.secret = config["api-secret"]
	front.store = sessions.NewCookieStore([]byte(front.secret))
	front.store.Options = &sessions.Options{
		// TODO: detect http:// etc in prefix
		Path:   "/",
		MaxAge: 600,
	}

	// liveui related members
	front.liveui_chnl = make(chan PostModel, 128)
	front.liveui_register = make(chan *liveChan, 128)
	front.liveui_deregister = make(chan *liveChan, 128)
	front.liveui_chans = make(map[string]*liveChan, 128)
	front.end_liveui = make(chan bool)
	return front
}
