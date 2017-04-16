package frontend

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"net/http"
	"nntpchan/lib/admin"
	"nntpchan/lib/api"
	"nntpchan/lib/config"
	"nntpchan/lib/database"
	"nntpchan/lib/model"
	"nntpchan/lib/nntp"
	"time"
)

// http frontend server
// provides glue layer between nntp and middleware
type httpFrontend struct {
	// bind address
	addr string
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
	db database.Database
}

// reload http frontend
// reloads middleware
func (f *httpFrontend) Reload(c *config.FrontendConfig) {
	if f.middleware == nil {
		if c.Middleware != nil {
			var err error
			// no middleware set, create middleware
			f.middleware, err = OverchanMiddleware(c.Middleware, f.db)
			if err != nil {
				log.Errorf("overchan middleware reload failed: %s", err.Error())
			}

		}
	} else {
		// middleware exists
		// do middleware reload
		f.middleware.Reload(c.Middleware)
	}

}

// serve http requests from net.Listener
func (f *httpFrontend) Serve() {
	// serve http
	for {
		err := http.ListenAndServe(f.addr, f.httpmux)
		if err != nil {
			log.Errorf("failed to listen and serve with frontend: %s", err)
		}
		time.Sleep(time.Second)
	}
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

func createHttpFrontend(c *config.FrontendConfig, mid Middleware, db database.Database) (f *httpFrontend, err error) {
	f = new(httpFrontend)
	// set db
	// db.Ensure() called elsewhere
	f.db = db

	// set bind address
	f.addr = c.BindAddr

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
