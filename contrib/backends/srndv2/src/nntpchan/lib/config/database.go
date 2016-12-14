package config

type DatabaseConfig struct {
	// url or address for database connector
	Addr string `json:"addr"`
	// password to use
	Password string `json:"password"`
	// username to use
	Username string `json:"username"`
	// type of database to use
	Type string `json:"type"`
}

var DefaultDatabaseConfig = DatabaseConfig{
	Type:     "postgres",
	Addr:     "/var/run/postgresql",
	Password: "",
}
