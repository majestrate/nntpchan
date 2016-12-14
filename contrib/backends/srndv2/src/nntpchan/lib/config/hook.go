package config

// config for external callback for nntp articles
type NNTPHookConfig struct {
	// name of hook
	Name string `json:"name"`
	// executable script path to be called with arguments: /path/to/article
	Exec string `json:"exec"`
}

// default dummy hook
var DefaultNNTPHookConfig = &NNTPHookConfig{
	Name: "dummy",
	Exec: "/bin/true",
}
