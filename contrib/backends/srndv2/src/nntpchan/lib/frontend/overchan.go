package frontend

import (
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/majestrate/srndv2/lib/cache"
	"github.com/majestrate/srndv2/lib/config"
	"github.com/majestrate/srndv2/lib/database"
	"html/template"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
)

const nntpchan_cache_key = "NNTPCHAN_CACHE::"

func cachekey(k string) string {
	return nntpchan_cache_key + k
}

func cachekey_for_thread(threadid string) string {
	return cachekey("thread-" + threadid)
}

func cachekey_for_board(name, page string) string {
	return cachekey("board-" + page + "-" + name)
}

// standard overchan imageboard middleware
type overchanMiddleware struct {
	templ       *template.Template
	markupCache cache.CacheInterface
	captcha     *CaptchaServer
	store       *sessions.CookieStore
	db          database.DB
}

func (m *overchanMiddleware) SetupRoutes(mux *mux.Router) {
	// setup front page handler
	mux.Path("/").HandlerFunc(m.ServeIndex)
	// setup thread handler
	mux.Path("/thread/{id}/").HandlerFunc(m.ServeThread)
	// setup board page handler
	mux.Path("/board/{name}/").HandlerFunc(m.ServeBoardPage)
	// setup posting endpoint
	mux.Path("/post")
	// create captcha
	captchaPrefix := "/captcha/"
	m.captcha = NewCaptchaServer(200, 400, captchaPrefix, m.store)
	// setup captcha endpoint
	m.captcha.SetupRoutes(mux.PathPrefix(captchaPrefix).Subrouter())
}

// reload middleware
func (m *overchanMiddleware) Reload(c *config.MiddlewareConfig) {
	// reload templates
	templ, err := template.ParseGlob(filepath.Join(c.Templates, "*.tmpl"))
	if err == nil {
		log.Infof("middleware reloaded templates")
		m.templ = templ
	} else {
		log.Errorf("middleware reload failed: %s", err.Error())
	}
}

func (m *overchanMiddleware) ServeBoardPage(w http.ResponseWriter, r *http.Request) {
	param := mux.Vars(r)
	board := param["name"]
	page := r.URL.Query().Get("page")
	if page == "" {
		page = "0"
	}
	pageno, err := strconv.Atoi(page)
	if err == nil {
		m.serveTemplate(w, r, "board.html.tmpl", cachekey_for_board(board, page), func() (interface{}, error) {
			// get object for cache miss
			// TODO: hardcoded page size
			return m.db.GetGroupForPage(board, pageno, 10)
		})
	} else {
		// 404
		http.NotFound(w, r)
	}
}

// serve cached thread
func (m *overchanMiddleware) ServeThread(w http.ResponseWriter, r *http.Request) {
	param := mux.Vars(r)
	thread_id := param["id"]
	m.serveTemplate(w, r, "thread.html.tmpl", cachekey_for_thread(thread_id), func() (interface{}, error) {
		// get object for cache miss
		return m.db.GetThreadByHash(thread_id)
	})
}

// serve index page
func (m *overchanMiddleware) ServeIndex(w http.ResponseWriter, r *http.Request) {
	m.serveTemplate(w, r, "index.html.tmpl", "index", nil)
}

// serve a template
func (m *overchanMiddleware) serveTemplate(w http.ResponseWriter, r *http.Request, tname, cacheKey string, getObj func() (interface{}, error)) {
	t := m.templ.Lookup(tname)
	if t == nil {
		log.WithFields(log.Fields{
			"template": tname,
		}).Warning("template not found")
		http.NotFound(w, r)
	} else {
		m.markupCache.ServeCached(w, r, cacheKey, func(wr io.Writer) error {
			if getObj == nil {
				return t.Execute(wr, nil)
			} else {
				// get model object
				obj, err := getObj()
				if err != nil {
					// error getting model
					log.WithFields(log.Fields{
						"error":    err,
						"template": tname,
						"cacheKey": cacheKey,
					}).Warning("failed to refresh template")
					return err
				}
				return t.Execute(wr, obj)
			}
		})
	}
}

// create standard overchan middleware
func OverchanMiddleware(c *config.MiddlewareConfig, markupCache cache.CacheInterface, db database.DB) (m Middleware, err error) {
	om := new(overchanMiddleware)
	om.markupCache = markupCache
	om.templ, err = template.ParseGlob(filepath.Join(c.Templates, "*.tmpl"))
	om.db = db
	if err == nil {
		m = om
	}
	return
}
