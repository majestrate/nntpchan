package config

// configuration for http middleware
type MiddlewareConfig struct {
	// middleware type, currently just 1 is available: overchan
	Type string `json:"type"`
	// directory for our html templates
	Templates string `json:"templates_dir"`
}

var DefaultMiddlewareConfig = MiddlewareConfig{
	Type:      "overchan",
	Templates: "./files/templates/overchan/",
}
