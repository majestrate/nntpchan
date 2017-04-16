package config

// configuration for a single web hook
type WebhookConfig struct {
	// user provided name for this hook
	Name string `json:"name"`
	// callback URL for webhook
	URL string `json:"url"`
	// dialect to use when calling webhook
	Dialect string `json:"dialect"`
}

var DefaultWebHookConfig = &WebhookConfig{
	Name:    "vichan",
	Dialect: "vichan",
	URL:     "http://localhost/webhook.php",
}
