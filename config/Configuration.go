package config

import (
	"encoding/json"
	"io/ioutil"
	"time"

	"github.com/gansoi/gansoi/checks"
	"github.com/gansoi/gansoi/database"
	"github.com/gansoi/gansoi/transports/ssh"
	"github.com/ghodss/yaml"
)

type (
	// Configuration keeps configuration for a core node.
	Configuration struct {
		Bind         string                   `json:"bind"`
		DataDir      string                   `json:"datadir"`
		HTTP         HTTP                     `json:"http"`
		HTTPRedirect HTTPRedirect             `json:"redirect"`
		Hosts        map[string]ssh.SSH       `json:"hosts"`
		Checks       map[string]*checks.Check `json:"checks"`
	}
)

const (
	// DefaultPath is the default location for the config file.
	DefaultPath = "/etc/gansoi.yml"
)

var (
	exampleConfig = `# Example configuration for gansoi.
bind: ":4934"
datadir: "/var/lib/gansoi"

http:
  bind: ":443"
  tls: true
  hostnames:
    - "gansoi.example.com"
  cert: "/etc/gansoi/me-cert.pem"
  key: "/etc/gansoi/me-key.pem"

redirect:
  bind: ":80"
  target: "https://gansoi.example.com/"
`
)

// SetDefaults sets some sane configuration defaults.
func (c *Configuration) SetDefaults() {
	// By default we bind to port 443 (HTTPS) on all interfaces on both IPv4
	// and IPv6.
	c.HTTP.Bind = ":443"
	c.HTTP.TLS = true

	c.HTTPRedirect.Bind = ":80"

	c.Bind = ":4934"

	// This makes sense on a unix system.
	c.DataDir = "/var/lib/gansoi"
}

// LoadFromFile loads a configuration from path.
func (c *Configuration) LoadFromFile(path string) error {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	return c.loadFromBytes(b)
}

func (c *Configuration) loadFromBytes(b []byte) error {
	err := yaml.Unmarshal(b, c)
	if err != nil {
		return err
	}

	// If the redirect target is empty, we default to the first hostname.
	if c.HTTPRedirect.Target == "" && len(c.HTTP.Hostnames) > 0 {
		c.HTTPRedirect.Target = scheme[c.HTTP.TLS] +
			"://" +
			c.HTTP.Hostnames[0] +
			"/"
	}

	return nil
}

// SaveChecks will save all checks from the configuration to the supplied
// database.Writer.
func (c *Configuration) SaveChecks(w database.Writer) error {
	for id, check := range c.Checks {
		if check.ID == "" {
			check.ID = id
		}

		if check.Name == "" {
			check.Name = id
		}

		if check.Arguments == nil {
			check.Arguments = json.RawMessage("{}")
		}

		check.ContactGroups = []string{}

		if check.Interval == 0 {
			check.Interval = time.Second * 30
		}

		err := w.Save(check)
		if err != nil {
			return err
		}
	}

	return nil
}

// SaveHosts will save all hosts from the configuration to the supplied
// database.Writer.
func (c *Configuration) SaveHosts(w database.Writer) error {
	for id, h := range c.Hosts {
		if h.ID == "" {
			h.ID = id
		}

		if h.Address == "" {
			h.Address = id
		}

		if h.Username == "" {
			h.Username = "root"
		}

		err := w.Save(&h)
		if err != nil {
			return err
		}
	}

	return nil
}
