package config

type (
	// HTTP configures the HTTP server.
	HTTP struct {
		TLS       bool     `toml:"tls"`
		Bind      string   `toml:"bind"`
		Hostnames []string `toml:"hostnames"`
		CertPath  string   `toml:"cert"`
		KeyPath   string   `toml:"key"`
		Login     string   `toml:"login"`
		Password  string   `toml:"password"`
	}
)
