//
// mod.go
// post moderation
//
package srnd

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

// regenerate pages function
type RegenFunc func(newsgroup, msgid, root string, page int)

// does an action for the administrator
// takes in json
type AdminFunc func(param map[string]interface{}) (interface{}, error)

// interface for moderation ui
type ModUI interface {

	// channel for daemon to poll for nntp articles from the mod ui
	MessageChan() chan NNTPMessage

	// check if this key is allowed to access
	// return true if it can otherwise false
	CheckKey(privkey, scope string) (bool, error)

	// serve the base page
	ServeModPage(wr http.ResponseWriter, r *http.Request)
	// handle a login POST request
	HandleLogin(wr http.ResponseWriter, r *http.Request)
	// handle a delete article request
	HandleDeletePost(wr http.ResponseWriter, r *http.Request)
	// handle a ban address request
	HandleBanAddress(wr http.ResponseWriter, r *http.Request)
	// handle an unban address request
	HandleUnbanAddress(wr http.ResponseWriter, r *http.Request)
	// handle add a pubkey
	HandleAddPubkey(wr http.ResponseWriter, r *http.Request)
	// handle removing a pubkey
	HandleDelPubkey(wr http.ResponseWriter, r *http.Request)
	// handle key generation
	HandleKeyGen(wr http.ResponseWriter, r *http.Request)
	// handle admin command
	HandleAdminCommand(wr http.ResponseWriter, r *http.Request)
}

type ModEvent interface {
	// turn it into a string for putting into an article
	String() string
	// what type of mod event
	Action() string
	// what reason for the event
	Reason() string
	// what is the event acting on
	Target() string
	// scope of the event, regex of newsgroup
	Scope() string
	// when this mod event expires, unix nano
	Expires() int64
}

type simpleModEvent string

func (self simpleModEvent) String() string {
	return string(self)
}

func (self simpleModEvent) Action() string {
	return strings.Split(string(self), " ")[0]
}

func (self simpleModEvent) Reason() string {
	return ""
}

func (self simpleModEvent) Target() string {
	return strings.Split(string(self), " ")[1]
}

func (self simpleModEvent) Scope() string {
	// TODO: hard coded
	return "overchan.*"
}

func (self simpleModEvent) Expires() int64 {
	// no expiration
	return -1
}

// create an overchan-delete mod event
func overchanDelete(msgid string) ModEvent {
	return simpleModEvent(fmt.Sprintf("delete %s", msgid))
}

// create an overchan-inet-ban mod event
func overchanInetBan(encAddr, key string, expire int64) ModEvent {
	return simpleModEvent(fmt.Sprintf("overchan-inet-ban %s:%s:%d", encAddr, key, expire))
}

// moderation message
// wraps multiple mod events
// is turned into an NNTPMessage later
type ModMessage []ModEvent

// write this mod message's body
func (self ModMessage) WriteTo(wr io.Writer, delim []byte) (err error) {
	// write body
	for _, ev := range self {
		_, err = io.WriteString(wr, ev.String())
		_, err = wr.Write(delim)
	}
	return
}

func ParseModEvent(line string) ModEvent {
	return simpleModEvent(line)
}

// wrap mod message in an nntp message
// does not sign
func wrapModMessage(mm ModMessage) NNTPMessage {
	pathname := "nntpchan.censor"
	nntp := &nntpArticle{
		headers: make(ArticleHeaders),
	}
	nntp.headers.Set("Newsgroups", "ctl")
	nntp.headers.Set("Content-Type", "text/plain; charset=UTF-8")
	nntp.headers.Set("Message-ID", genMessageID(pathname))
	nntp.headers.Set("Date", timeNowStr())
	nntp.headers.Set("Path", pathname)
	// todo: set these maybe?
	nntp.headers.Set("From", "anon <a@n.on>")
	nntp.headers.Set("Subject", "censor")

	var buff bytes.Buffer
	// crlf delimited
	_ = mm.WriteTo(&buff, []byte{10})
	// create plaintext attachment, cut off last 2 bytes
	nntp.message = buff.String()
	buff.Reset()
	return nntp
}

type ModEngine interface {
	// chan to send the mod engine posts given message_id
	MessageChan() chan string
	// delete post of a poster
	DeletePost(msgid string, regen RegenFunc) error
	// ban a cidr
	BanAddress(cidr string) error
	// do we allow this public key to delete this message-id ?
	AllowDelete(pubkey, msgid string) bool
	// do we allow this public key to do inet-ban?
	AllowBan(pubkey string) bool
	// load a mod message
	LoadMessage(msgid string) NNTPMessage
}

type modEngine struct {
	database Database
	store    ArticleStore
	chnl     chan string
}

func (self modEngine) LoadMessage(msgid string) NNTPMessage {
	return self.store.GetMessage(msgid)
}

func (self modEngine) MessageChan() chan string {
	return self.chnl
}

func (self modEngine) BanAddress(cidr string) (err error) {
	return self.database.BanAddr(cidr)
}

func (self modEngine) DeletePost(msgid string, regen RegenFunc) (err error) {
	hdr, err := self.database.GetHeadersForMessage(msgid)
	var delposts []string
	var page int64
	var ref, group string
	rootmsgid := ""
	if hdr == nil {
		log.Println("failed to get headers for article", msgid, err)
	} else {
		ref = hdr.Get("References", "")
		group = hdr.Get("Newsgroups", "")
		if ref == "" || ref == msgid {
			// is root post
			// delete replies too
			repls := self.database.GetThreadReplies(msgid, 0, 0)
			if repls == nil {
				log.Println("cannot get thread replies for", msgid)
			} else {
				delposts = append(delposts, repls...)
			}

			_, page, err = self.database.GetPageForRootMessage(msgid)
			ref = msgid
			rootmsgid = msgid
		} else {
			_, page, err = self.database.GetPageForRootMessage(ref)
		}
	}
	delposts = append(delposts, msgid)
	// get list of files to delete
	var delfiles []string
	for _, delmsg := range delposts {
		article := self.store.GetFilename(delmsg)
		delfiles = append(delfiles, article)
		// get attachments for post
		atts := self.database.GetPostAttachments(delmsg)
		if atts != nil {
			for _, att := range atts {
				img := self.store.AttachmentFilepath(att)
				thm := self.store.ThumbnailFilepath(att)
				delfiles = append(delfiles, img, thm)
			}
		}
	}
	// delete all files
	for _, f := range delfiles {
		log.Printf("delete file: %s", f)
		os.Remove(f)
	}

	if rootmsgid != "" {
		self.database.DeleteThread(rootmsgid)
	}

	for _, delmsg := range delposts {
		// delete article from post database
		err = self.database.DeleteArticle(delmsg)
		if err != nil {
			log.Println(err)
		}
		// ban article
		self.database.BanArticle(delmsg, "deleted by moderator")
	}
	regen(group, msgid, ref, int(page))
	return nil
}

func (self modEngine) AllowBan(pubkey string) bool {
	is_admin, _ := self.database.CheckAdminPubkey(pubkey)
	if is_admin {
		// admins can do whatever
		return true
	}
	return self.database.CheckModPubkeyGlobal(pubkey)
}

func (self modEngine) AllowDelete(pubkey, msgid string) (allow bool) {
	is_admin, _ := self.database.CheckAdminPubkey(pubkey)
	if is_admin {
		// admins can do whatever
		return true
	}
	if self.database.CheckModPubkeyGlobal(pubkey) {
		// globals can delete as they wish
		return true
	}
	// check for scoped permissions
	_, group, _, err := self.database.GetInfoForMessage(msgid)
	if err == nil && newsgroupValidFormat(group) {
		allow = self.database.CheckModPubkeyCanModGroup(pubkey, group)
	} else if err != nil {
		log.Println("db error in mod engine while checking permissions", err)
	}
	return
}

// run a mod engine logic mainloop
func RunModEngine(mod ModEngine, regen RegenFunc) {

	chnl := mod.MessageChan()
	for {
		msgid := <-chnl
		nntp := mod.LoadMessage(msgid)
		if nntp == nil {
			log.Println("failed to load mod message", msgid)
			continue
		}
		// sanity check
		if nntp.Newsgroup() == "ctl" {
			pubkey := nntp.Pubkey()
			for _, line := range strings.Split(nntp.Message(), "\n") {
				line = strings.Trim(line, "\r\t\n ")
				ev := ParseModEvent(line)
				action := ev.Action()
				if action == "delete" {
					msgid := ev.Target()
					if !ValidMessageID(msgid) {
						// invalid message-id
						log.Println("invalid message-id for mod delete", msgid, "from", pubkey)
						continue
					}
					// this is a delete action
					if mod.AllowDelete(pubkey, msgid) {
						err := mod.DeletePost(msgid, regen)
						if err != nil {
							log.Println(msgid, err)
						}
					} else {
						log.Printf("pubkey=%s will not delete %s not trusted", pubkey, msgid)
					}
				} else if action == "overchan-inet-ban" {
					// ban action
					target := ev.Target()
					if target[0] == '[' {
						// probably a literal ipv6 rangeban
						if mod.AllowBan(pubkey) {
							err := mod.BanAddress(target)
							if err != nil {
								log.Println("failed to do literal ipv6 range ban on", target, err)
							}
						} else {
							log.Println("ignoring literal ipv6 rangeban from", pubkey, "as they are not allowed to ban")
						}
						continue
					}
					parts := strings.Split(target, ":")
					if len(parts) == 3 {
						// encrypted ip
						encaddr, key := parts[0], parts[1]
						cidr := decAddr(encaddr, key)
						if cidr == "" {
							log.Println("failed to decrypt inet ban")
						} else if mod.AllowBan(pubkey) {
							err := mod.BanAddress(cidr)
							if err != nil {
								log.Println("failed to do range ban on", cidr, err)
							}
						} else {
							log.Println("ingoring encrypted-ip inet ban from", pubkey, "as they are not allowed to ban")
						}
					} else if len(parts) == 1 {
						// literal cidr
						cidr := parts[0]
						if mod.AllowBan(pubkey) {
							err := mod.BanAddress(cidr)
							if err != nil {
								log.Println("failed to do literal range ban on", cidr, err)
							}
						} else {
							log.Println("ingoring literal cidr range ban from", pubkey, "as they are not allowed to ban")
						}
					} else {
						log.Printf("invalid overchan-inet-ban: target=%s", target)
					}
				} else {
					log.Println("invalid mod action", action, "from", pubkey)
				}
			}
		}
	}
}
