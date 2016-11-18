package api

import (
	"github.com/gorilla/mux"
	"net/http"
)

// api server
type Server struct {
}

func (s *Server) HandlePing(w http.ResponseWriter, r *http.Request) {

}

// inject api routes
func (s *Server) SetupRoutes(r *mux.Router) {
	// setup api pinger
	r.Path("/ping").HandlerFunc(s.HandlePing)
}
