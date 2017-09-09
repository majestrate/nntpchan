package config

import (
	"fmt"
)

type FrontendConfig struct {
	// bind to address
	BindAddr string `json:"bind"`
	// frontend cache
	Cache *CacheConfig `json:"cache"`
	// frontend ssl settings
	SSL *SSLSettings `json:"ssl"`
	// static files directory
	Static string `json:"static_dir"`
	// http middleware configuration
	Middleware *MiddlewareConfig `json:"middleware"`
	// storage config
	Storage *StoreConfig `json:"-"`
}

func (cfg *FrontendConfig) Name() string {
	return fmt.Sprintf("frontend-%s", cfg.BindAddr)
}

// default Frontend Configuration
var DefaultFrontendConfig = FrontendConfig{
	BindAddr:   "127.0.0.1:18888",
	Static:     "./files/static/",
	Middleware: &DefaultMiddlewareConfig,
}
