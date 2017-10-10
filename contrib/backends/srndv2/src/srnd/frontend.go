//
// frontend.go
// srnd frontend interfaces
//
//
package srnd

const BumpLimit = 300

// ( message-id, references, newsgroup )
type frontendPost [3]string

func (p frontendPost) MessageID() string {
	return p[0]
}

func (p frontendPost) Reference() string {
	return p[1]
}

func (p frontendPost) Newsgroup() string {
	return p[2]
}

// frontend interface for any type of frontend
type Frontend interface {

	// handle new post from nntpd
	HandleNewPost(p frontendPost)

	// run mainloop
	Mainloop()

	// do we want posts from a newsgroup?
	AllowNewsgroup(group string) bool

	// trigger a manual regen of indexes for a root post
	Regen(msg ArticleEntry)

	// regenerate on mod event
	RegenOnModEvent(newsgroup, msgid, root string, page int)

	GetCacheHandler() CacheHandler
}
