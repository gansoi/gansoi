package config

import "github.com/BurntSushi/toml"

type (
	// Configuration keeps configuration for a core node.
	Configuration struct {
		Bind    string `toml:"bind"`
		DataDir string `toml:"datadir"`
		HTTP    HTTP   `toml:"http"`
	}
)

const (
	// DefaultPath is the default location for the config file.
	DefaultPath = "/etc/gansoi.conf"
)

var (
	exampleConfig = `# Example configuration for gansoi.
bind = ":4934"
datadir = "/var/lib/gansoi"

[http]
bind = ":443"
tls = true
hostnames = [ "gansoi.example.com" ]
cert = "/etc/gansoi/me-cert.pem"
key = "/etc/gansoi/me-key.pem"
`
)

// SetDefaults sets some sane configuration defaults.
func (c *Configuration) SetDefaults() {
	// By default we bind to port 443 (HTTPS) on all interfaces on both IPv4
	// and IPv6.
	c.HTTP.Bind = ":443"
	c.HTTP.TLS = true

	c.Bind = ":4934"

	// This makes sense on a unix system.
	c.DataDir = "/var/lib/gansoi"
}

// LoadFromFile loads a configuration from path.
func (c *Configuration) LoadFromFile(path string) error {
	_, err := toml.DecodeFile(path, c)

	return err
}
