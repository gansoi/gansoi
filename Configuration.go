package main

import (
	"net"
	"net/url"
	"os"
	"sync"

	"github.com/BurntSushi/toml"
)

type (
	// Configuration keeps configuration for a core node.
	Configuration struct {
		sync.RWMutex
		BindPrivate string `toml:"private"`
		BindPublic  string `toml:"public"`
		Cert        string `toml:"cert"`
		Key         string `toml:"key"`
		DataDir     string `toml:"datadir"`
		Hostname    string `toml:"hostname"`
		Secret      string `toml:"secret"`
		LetsEncrypt bool   `toml:"letsencrypt"`
		Login       string `toml:"login"`
		Password    string `toml:"password"`
	}
)

var (
	configFile = "/etc/gansoi.conf"

	exampleConfig = `# Example configuration for gansoi.
private = ":4934"
public = "https://0.0.0.0/:443"
cert = "/etc/gansoi/me-cert.pem"
key = "/etc/gansoi/me-key.pem"
datadir = "/var/lib/gansoi"
hostname = "gansoi.example.com"

# cert and key are ignored if set
letsencrypt = true
`
)

// SetDefaults sets some sane configuration defaults.
func (c *Configuration) SetDefaults() {
	// By default we bind to port 443 (HTTPS) on all interfaces on both IPv4
	// and IPv6.
	c.BindPublic = ":443"

	c.BindPrivate = ":4934"

	// This makes sense on a unix system.
	c.DataDir = "/var/lib/gansoi"
}

// LoadFromFile loads a configuration from path.
func (c *Configuration) LoadFromFile(path string) error {
	c.Lock()
	defer c.Unlock()

	_, err := toml.DecodeFile(path, c)

	return err
}

// Bind will return a string suitable for binding.
func (c *Configuration) Bind() string {
	URL, _ := url.Parse(c.BindPublic)

	host, port, err := net.SplitHostPort(URL.Host)
	if err != nil {
		return ":443"
	}

	if port == "" {
		switch URL.Scheme {
		case "https":
			port = "443"
		case "http":
			port = "80"
		}
	}

	return net.JoinHostPort(host, port)
}

// Hostnames returns a list of hostnames exposed.
func (c *Configuration) Hostnames() []string {
	c.RLock()
	defer c.RUnlock()

	hostname, _ := os.Hostname()

	hostnames := []string{hostname}

	local, _, _ := net.SplitHostPort(c.BindPublic)
	if local != "" {
		hostnames = append(hostnames, local)
	}

	if c.Hostname != "" {
		hostnames = append(hostnames, c.Hostname)
	}

	return hostnames
}

// TLS will return true if we should set up TLS.
func (c *Configuration) TLS() bool {
	URL, _ := url.Parse(c.BindPublic)

	return URL.Scheme == "https"
}
