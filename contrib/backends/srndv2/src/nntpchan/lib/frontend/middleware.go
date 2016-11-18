package frontend

import (
	"github.com/gorilla/mux"
	"github.com/majestrate/srndv2/lib/config"
)

// http middleware
type Middleware interface {
	// set up routes
	SetupRoutes(m *mux.Router)
	// reload with new configuration
	Reload(c *config.MiddlewareConfig)
}
