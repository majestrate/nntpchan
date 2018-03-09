package config

// configuration for http middleware
type MiddlewareConfig struct {
	// middleware type, currently just 1 is available: overchan
	Type string `json:"type"`
	// directory for our html templates
	Templates string `json:"templates_dir"`
	// directory for static files
	StaticDir string `json:"static_dir"`
}

var DefaultMiddlewareConfig = MiddlewareConfig{
	Type:      "overchan",
	Templates: "./files/templates/overchan/",
	StaticDir: "./files/",
}
