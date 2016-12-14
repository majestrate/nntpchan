package frontend

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/majestrate/srndv2/lib/admin"
	"github.com/majestrate/srndv2/lib/api"
	"github.com/majestrate/srndv2/lib/cache"
	"github.com/majestrate/srndv2/lib/config"
	"github.com/majestrate/srndv2/lib/database"
	"github.com/majestrate/srndv2/lib/model"
	"github.com/majestrate/srndv2/lib/nntp"
	"net"
	"net/http"
)

// http frontend server
// provides glue layer between nntp and middleware
type httpFrontend struct {
	// http mux
	httpmux *mux.Router
	// admin panel
	adminPanel *admin.Server
	// static files path
	staticDir string
	// http middleware
	middleware Middleware
	// api server
	apiserve *api.Server
	// database driver
	db database.DB
}

// reload http frontend
// reloads middleware
func (f *httpFrontend) Reload(c *config.FrontendConfig) {
	if f.middleware == nil {
		if c.Middleware != nil {
			markupcache, err := cache.FromConfig(c.Cache)
			if err == nil {
				// no middleware set, create middleware
				f.middleware, err = OverchanMiddleware(c.Middleware, markupcache, f.db)
				if err != nil {
					log.Errorf("overchan middleware reload failed: %s", err.Error())
				}
			} else {
				// error creating cache
				log.Errorf("failed to create cache: %s", err.Error())
			}
		}
	} else {
		// middleware exists
		// do middleware reload
		f.middleware.Reload(c.Middleware)
	}

}

// serve http requests from net.Listener
func (f *httpFrontend) Serve(l net.Listener) (err error) {
	// serve http
	err = http.Serve(l, f.httpmux)
	return
}

// serve robots.txt page
func (f *httpFrontend) serveRobots(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "User-Agent: *\nDisallow: /\n")
}

func (f *httpFrontend) AllowPost(p model.PostReference) bool {
	// TODO: implement
	return true
}

func (f *httpFrontend) Regen(p model.PostReference) {
	// TODO: implement
}

func (f *httpFrontend) GotArticle(msgid nntp.MessageID, group nntp.Newsgroup) {
	// TODO: implement
}

func (f *httpFrontend) SentArticleVia(msgid nntp.MessageID, feedname string) {
	// TODO: implement
}

func createHttpFrontend(c *config.FrontendConfig, mid Middleware, db database.DB) (f *httpFrontend, err error) {
	f = new(httpFrontend)
	// set db
	// db.Ensure() called elsewhere
	f.db = db

	// set up mux
	f.httpmux = mux.NewRouter()

	// set up admin panel
	f.adminPanel = admin.NewServer()

	// set static files dir
	f.staticDir = c.Static

	// set middleware
	f.middleware = mid

	// set up routes

	if f.adminPanel != nil {
		// route up admin panel
		f.httpmux.PathPrefix("/admin/").Handler(f.adminPanel)
	}

	if f.middleware != nil {
		// route up middleware
		f.middleware.SetupRoutes(f.httpmux)
	}

	if f.apiserve != nil {
		// route up api
		f.apiserve.SetupRoutes(f.httpmux.PathPrefix("/api/").Subrouter())
	}

	// route up robots.txt
	f.httpmux.Path("/robots.txt").HandlerFunc(f.serveRobots)

	// route up static files
	f.httpmux.PathPrefix("/static/").Handler(http.FileServer(http.Dir(f.staticDir)))

	return
}
