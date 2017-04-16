package frontend

import (
	"github.com/gorilla/mux"
	"nntpchan/lib/config"
)

// http middleware
type Middleware interface {
	// set up routes
	SetupRoutes(m *mux.Router)
	// reload with new configuration
	Reload(c *config.MiddlewareConfig)
}
