package admin

import (
	"net/http"
)

type Server struct {
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {

}

func NewServer() *Server {
	return &Server{}
}
