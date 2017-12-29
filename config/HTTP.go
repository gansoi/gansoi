package config

type (
	// HTTP configures the HTTP server.
	HTTP struct {
		TLS       bool     `json:"tls"`
		Bind      string   `json:"bind"`
		Hostnames []string `json:"hostnames"`
		CertPath  string   `json:"cert"`
		KeyPath   string   `json:"key"`
		Login     string   `json:"login"`
		Password  string   `json:"password"`
	}
)

var (
	scheme = map[bool]string{
		false: "http",
		true:  "https",
	}
)
