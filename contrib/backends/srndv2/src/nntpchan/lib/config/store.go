package config

type StoreConfig struct {
	// path to article directory
	Path string `json:"path"`
}

var DefaultStoreConfig = StoreConfig{
	Path: "storage",
}
