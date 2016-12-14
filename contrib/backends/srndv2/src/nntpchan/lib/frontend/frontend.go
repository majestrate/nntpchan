package frontend

import (
	"github.com/majestrate/srndv2/lib/cache"
	"github.com/majestrate/srndv2/lib/config"
	"github.com/majestrate/srndv2/lib/database"
	"github.com/majestrate/srndv2/lib/model"
	"github.com/majestrate/srndv2/lib/nntp"

	"net"
)

// a frontend that displays nntp posts and allows posting
type Frontend interface {

	// run mainloop using net.Listener
	Serve(l net.Listener) error

	// do we accept this inbound post?
	AllowPost(p model.PostReference) bool

	// trigger a manual regen of indexes for a root post
	Regen(p model.PostReference)

	// implements nntp.EventHooks
	GotArticle(msgid nntp.MessageID, group nntp.Newsgroup)

	// implements nntp.EventHooks
	SentArticleVia(msgid nntp.MessageID, feedname string)

	// reload config
	Reload(c *config.FrontendConfig)
}

// create a new http frontend give frontend config
func NewHTTPFrontend(c *config.FrontendConfig, db database.DB) (f Frontend, err error) {

	var markupCache cache.CacheInterface

	markupCache, err = cache.FromConfig(c.Cache)
	if err != nil {
		return
	}

	var mid Middleware
	if c.Middleware != nil {
		// middleware configured
		mid, err = OverchanMiddleware(c.Middleware, markupCache, db)
	}

	if err == nil {
		// create http frontend only if no previous errors
		f, err = createHttpFrontend(c, mid, db)
	}
	return
}
