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
		Local       string   `toml:"local"`
		Cert        string   `toml:"cert"`
		Key         string   `toml:"key"`
		DataDir     string   `toml:"datadir"`
		Cluster     []string `toml:"cluster"`
		ClusterName string   `toml:"cluster_name"`
		Secret      string   `toml:"secret"`
		LetsEncrypt bool     `toml:"letsencrypt"`
		self        string
		Login       string `toml:"login"`
		Password    string `toml:"password"`
	}
)

var (
	configFile = "/etc/gansoi.conf"

	exampleConfig = `# Example configuration for gansoi.
local = "london.example.com"
cert = "/etc/gansoi/me-cert.pem"
key = "/etc/gansoi/me-key.pem"
datadir = "/var/lib/gansoi"
cluster = ["london.example.com", "copenhagen.example.com", "berlin.example.com"]
cluster_name = "gansoi.example.com"
secret = "This is unsecure. Pick a good alphanumeric secret."

# cert and key are ignored if set
letsencrypt = true
`
)

// SetDefaults sets some sane configuration defaults.
func (c *Configuration) SetDefaults() {
	// By default we bind to port 443 (HTTPS) on all interfaces on both IPv4
	// and IPv6.
	c.Local = ":443"

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

// Self returns a string that can be used to describe this node.
func (c *Configuration) Self() string {
	c.Lock()
	defer c.Unlock()

	if c.self != "" {
		return c.self
	}

	if c.Local == "" {
		hostname, _ := os.Hostname()
		c.self = hostname
	}

	c.self = c.Local

	return c.self
}

// Bind will return a string suitable for binding.
func (c *Configuration) Bind() string {
	URL, _ := url.Parse(c.Local)

	host, port, err := net.SplitHostPort(URL.Host)
	if err != nil {
		return c.Local
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

	local, _, _ := net.SplitHostPort(c.Local)
	if local != "" {
		hostnames = append(hostnames, local)
	}

	if c.ClusterName != "" {
		hostnames = append(hostnames, c.ClusterName)
	}

	return hostnames
}

// TLS will return true if we should set up TLS.
func (c *Configuration) TLS() bool {
	URL, _ := url.Parse(c.Local)

	return URL.Scheme == "https"
}
