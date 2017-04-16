package frontend

import (
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"html/template"
	"net/http"
	"nntpchan/lib/config"
	"nntpchan/lib/database"
	"path/filepath"
	"strconv"
)

// standard overchan imageboard middleware
type overchanMiddleware struct {
	templ   *template.Template
	captcha *CaptchaServer
	store   *sessions.CookieStore
	db      database.Database
}

func (m *overchanMiddleware) SetupRoutes(mux *mux.Router) {
	// setup front page handler
	mux.Path("/").HandlerFunc(m.ServeIndex)
	// setup thread handler
	mux.Path("/t/{id}/").HandlerFunc(m.ServeThread)
	// setup board page handler
	mux.Path("/b/{name}/").HandlerFunc(m.ServeBoardPage)
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
	page := r.URL.Query().Get("q")
	pageno, err := strconv.Atoi(page)
	if err == nil {
		var obj interface{}
		obj, err = m.db.BoardPage(board, pageno, 10)
		if err == nil {
			m.serveTemplate(w, r, "board.html.tmpl", obj)
		} else {
			m.serveTemplate(w, r, "error.html.tmpl", err)
		}
	} else {
		// 404
		http.NotFound(w, r)
	}
}

// serve cached thread
func (m *overchanMiddleware) ServeThread(w http.ResponseWriter, r *http.Request) {
	param := mux.Vars(r)
	obj, err := m.db.ThreadByHash(param["id"])
	if err == nil {
		m.serveTemplate(w, r, "thread.html.tmpl", obj)
	} else {
		m.serveTemplate(w, r, "error.html.tmpl", err)
	}
}

// serve index page
func (m *overchanMiddleware) ServeIndex(w http.ResponseWriter, r *http.Request) {
	m.serveTemplate(w, r, "index.html.tmpl", nil)
}

// serve a template
func (m *overchanMiddleware) serveTemplate(w http.ResponseWriter, r *http.Request, tname string, obj interface{}) {
	t := m.templ.Lookup(tname)
	if t == nil {
		log.WithFields(log.Fields{
			"template": tname,
		}).Warning("template not found")
		http.NotFound(w, r)
	} else {
		err := t.Execute(w, obj)
		if err != nil {
			// error getting model
			log.WithFields(log.Fields{
				"error":    err,
				"template": tname,
			}).Warning("failed to render template")
		}
	}
}

// create standard overchan middleware
func OverchanMiddleware(c *config.MiddlewareConfig, db database.Database) (m Middleware, err error) {
	om := new(overchanMiddleware)
	om.templ, err = template.ParseGlob(filepath.Join(c.Templates, "*.tmpl"))
	om.db = db
	if err == nil {
		m = om
	}
	return
}
