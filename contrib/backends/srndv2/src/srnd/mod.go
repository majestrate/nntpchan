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
	// get outbound message channel
	MessageChan() chan NNTPMessage
}

type ModAction string

const ModInetBan = ModAction("overchan-inet-ban")
const ModDelete = ModAction("delete")
const ModRemoveAttachment = ModAction("overchan-del-attachment")
const ModStick = ModAction("overchan-stick")
const ModLock = ModAction("overchan-lock")
const ModHide = ModAction("overchan-hide")
const ModSage = ModAction("overchan-sage")
const ModDeleteAlt = ModAction("delete")

type ModEvent interface {
	// turn it into a string for putting into an article
	String() string
	// what type of mod event
	Action() ModAction
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

func (self simpleModEvent) Action() ModAction {
	switch strings.Split(string(self), " ")[0] {
	case "delete":
		return ModDelete
	case "overchan-inet-ban":
		return ModInetBan
	}
	return ""
}

func (self simpleModEvent) Reason() string {
	return ""
}

func (self simpleModEvent) Target() string {
	parts := strings.Split(string(self), " ")
	if len(parts) > 1 {
		return parts[1]
	}
	return ""
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
	// load and handle a mod message from ctl after it's verified
	HandleMessage(msgid string)
	// delete post of a poster
	DeletePost(msgid string) error
	// ban a cidr
	BanAddress(cidr string) error
	// do we allow this public key to delete this message-id ?
	AllowDelete(pubkey, msgid string) bool
	// do we allow this public key to do inet-ban?
	AllowBan(pubkey string) bool
	// allow janitor
	AllowJanitor(pubkey string) bool
	// load a mod message
	LoadMessage(msgid string) NNTPMessage
	// execute 1 mod action line by a mod with pubkey
	Execute(ev ModEvent, pubkey string)
	// do a mod event unconditionally
	Do(ev ModEvent)
}

type modEngine struct {
	database Database
	store    ArticleStore
	regen    RegenFunc
}

func (self *modEngine) LoadMessage(msgid string) NNTPMessage {
	return self.store.GetMessage(msgid)
}

func (self *modEngine) BanAddress(cidr string) (err error) {
	return self.database.BanAddr(cidr)
}

func (self *modEngine) DeletePost(msgid string) (err error) {
	hdr := self.store.GetHeaders(msgid)
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

	for _, delmsg := range delposts {
		// delete article from post database
		err = self.database.DeleteArticle(delmsg)
		if err != nil {
			log.Println(err)
		}
		// ban article
		self.database.BanArticle(delmsg, "deleted by moderator")
		self.store.Remove(delmsg)
	}

	if rootmsgid != "" {
		self.database.DeleteThread(rootmsgid)
	}
	self.regen(group, msgid, ref, int(page))
	return nil
}

func (self *modEngine) AllowBan(pubkey string) bool {
	is_admin, _ := self.database.CheckAdminPubkey(pubkey)
	if is_admin {
		// admins can do whatever
		return true
	}
	return self.database.CheckModPubkeyGlobal(pubkey)
}

func (self *modEngine) AllowJanitor(pubkey string) bool {
	is_admin, _ := self.database.CheckAdminPubkey(pubkey)
	if is_admin {
		return true
	}
	if self.database.CheckModPubkeyGlobal(pubkey) {
		return true
	}
	// TODO: more checks
	return false
}

func (self *modEngine) AllowDelete(pubkey, msgid string) (allow bool) {
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

func (mod *modEngine) HandleMessage(msgid string) {
	nntp := mod.store.GetMessage(msgid)
	if nntp == nil {
		log.Println("failed to load", msgid, "in mod engine, missing message")
		return
	}
	// sanity check
	if nntp.Newsgroup() == "ctl" {
		pubkey := nntp.Pubkey()
		for _, line := range strings.Split(nntp.Message(), "\n") {
			line = strings.Trim(line, "\r\t\n ")
			if len(line) > 0 {
				ev := ParseModEvent(line)
				mod.Execute(ev, pubkey)
			}
		}
	}
}

func (mod *modEngine) Do(ev ModEvent) {
	action := ev.Action()
	target := ev.Target()
	if action == ModDelete || action == ModDeleteAlt {
		msgid := target
		if !ValidMessageID(msgid) {
			// invalid message-id
			log.Println("invalid message-id", msgid)
			return
		}
		err := mod.DeletePost(msgid)
		if err != nil {
			log.Println(msgid, err)
		} else {
			log.Println("deleted", msgid)
		}

	} else if action == ModInetBan {
		// ban action
		if target[0] == '[' {
			err := mod.BanAddress(target)
			if err != nil {
				log.Println("failed to do literal ipv6 range ban on", target, err)
			} else {
				log.Println("banned", target)
			}
			return
		}
		parts := strings.Split(target, ":")
		if len(parts) == 3 {
			// encrypted ip
			encaddr, key := parts[0], parts[1]
			cidr := decAddr(encaddr, key)
			if cidr == "" {
				log.Println("failed to decrypt inet ban")
			} else {
				err := mod.BanAddress(cidr)
				if err != nil {
					log.Println("failed to do range ban on", cidr, err)
				} else {
					log.Println("banned", cidr)
				}
			}
		} else if len(parts) == 2 {
			// x-encrypted-ip ban without pad
			err := mod.database.BanEncAddr(parts[0])
			if err != nil {
				log.Println("failed to ban encrypted ip", err)
			} else {
				log.Println("banned poster", parts[0])
			}

		} else if len(parts) == 1 {
			// literal cidr
			cidr := parts[0]
			err := mod.BanAddress(cidr)
			if err != nil {
				log.Println("failed to do literal range ban on", cidr, err)
			} else {
				log.Println("banned cidr", cidr)
			}

		} else {
			log.Printf("invalid overchan-inet-ban: target=%s", target)
		}
	} else if action == ModHide {
		// TODO: implement
	} else if action == ModLock {
		// TODO: implement
	} else if action == ModSage {
		// TODO: implement
	} else if action == ModStick {
		// TODO: implement
	} else if action == ModRemoveAttachment {
		var delfiles []string
		atts := mod.database.GetPostAttachments(target)
		if atts != nil {
			for _, att := range atts {
				delfiles = append(delfiles, mod.store.AttachmentFilepath(att))
				delfiles = append(delfiles, mod.store.ThumbnailFilepath(att))
			}
		}
		for _, f := range delfiles {
			log.Println("remove file", f)
			os.Remove(f)
		}
	} else {
		log.Println("invalid mod action", action)
	}
}

func (mod *modEngine) Execute(ev ModEvent, pubkey string) {
	action := ev.Action()
	target := ev.Target()
	switch action {
	case ModDelete:
		if mod.AllowDelete(pubkey, target) {
			mod.Do(ev)
		}
		return
	case ModInetBan:
		if mod.AllowBan(pubkey) {
			mod.Do(ev)
		}
		return
	case ModHide:
	case ModLock:
	case ModSage:
	case ModStick:
	case ModRemoveAttachment:
		if mod.AllowJanitor(pubkey) {
			mod.Do(ev)
		}
	default:
		// invalid action
	}
}
