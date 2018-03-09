//
// expiration.go
// content expiration
//
package srnd

import (
	"log"
	"os"
	"path/filepath"
	"time"
)

// content expiration interface
type ExpirationCore interface {
	// do expiration for a group
	ExpireGroup(newsgroup string, keep int)
	// Delete a single post and all children
	ExpirePost(messageID string)
	// expire all orphaned articles
	ExpireOrphans()
	// expire all articles posted before time
	ExpireBefore(t time.Time)
}

type ExpireCacheFunc func(string, string, string)

func createExpirationCore(database Database, store ArticleStore, ex ExpireCacheFunc) ExpirationCore {
	return expire{database, store, ex}
}

type deleteEvent string

func (self deleteEvent) Path() string {
	return string(self)
}

func (self deleteEvent) MessageID() string {
	return filepath.Base(string(self))
}

type expire struct {
	database    Database
	store       ArticleStore
	expireCache ExpireCacheFunc
}

func (self expire) ExpirePost(messageID string) {
	headers := self.store.GetHeaders(messageID)
	// get article headers
	if headers != nil {
		group := headers.Get("Newsgroups", "")
		// is this a root post ?
		ref := headers.Get("References", "")
		if ref == "" || ref == messageID {
			// ya, expire the entire thread
			self.ExpireThread(group, messageID)
		} else {
			self.handleEvent(deleteEvent(self.store.GetFilename(messageID)))
			self.expireCache(group, messageID, ref)
		}
	}
}

func (self expire) ExpireGroup(newsgroup string, keep int) {
	threads := self.database.GetRootPostsForExpiration(newsgroup, keep)
	for _, root := range threads {
		self.ExpireThread(newsgroup, root)
	}
}

func (self expire) ExpireThread(group, rootMsgid string) {
	files, err := self.database.GetThreadAttachments(rootMsgid)
	if err == nil {
		for _, file := range files {
			img := self.store.AttachmentFilepath(file)
			os.Remove(img)
			thm := self.store.ThumbnailFilepath(file)
			os.Remove(thm)
		}
	} else {
		log.Println("expirethread::GetThreadAttachments:", err)
	}

	replies := self.database.GetThreadReplies(rootMsgid, 0, 0)

	for _, msgid := range replies {
		self.store.Remove(msgid)
	}

	self.store.Remove(rootMsgid)
	self.database.DeleteThread(rootMsgid)
	self.expireCache(group, rootMsgid, rootMsgid)
}

func (self expire) ExpireBefore(t time.Time) {
	articles, err := self.database.GetPostsBefore(t)
	if err == nil {
		for _, msgid := range articles {
			self.ExpirePost(msgid)
		}
	} else {
		log.Println("failed to expire older posts", err)
	}
}

// expire all orphaned articles
func (self expire) ExpireOrphans() {
	// get all articles in database
	articles := self.database.GetAllArticles()
	if articles != nil {
		log.Println("expire all orphan posts")
		// for each article
		for _, article := range articles {
			// load headers
			hdr := self.store.GetHeaders(article.MessageID())
			if hdr == nil {
				// article does not exist?
				// ensure it's deleted
				self.ExpirePost(article.MessageID())
			} else {
				// check if we are a reply
				rootMsgid := hdr.Get("References", "")
				if len(rootMsgid) == 0 {
					// root post
				} else {
					// reply
					// do we have this root post?
					if self.store.HasArticle(rootMsgid) {
						// yes, do nothing
					} else {
						// no, expire post
						self.ExpirePost(article.MessageID())
					}
				}
			}
		}
	}
}

func (self expire) handleEvent(ev deleteEvent) {
	atts := self.database.GetPostAttachments(ev.MessageID())
	// remove all attachments
	if atts != nil {
		for _, att := range atts {
			img := self.store.AttachmentFilepath(att)
			os.Remove(img)
			thm := self.store.ThumbnailFilepath(att)
			os.Remove(thm)
		}
	}
	banned := self.database.ArticleBanned(ev.MessageID())
	if !banned {
		err := self.database.BanArticle(ev.MessageID(), "expired")
		if err != nil {
			log.Println("failed to ban for expiration", err)
		}
	}
	err := self.database.DeleteArticle(ev.MessageID())
	if err != nil {
		log.Println("failed to delete article", err)
	}
	// remove article
	self.store.Remove(ev.MessageID())
}
