package config

import (
	"encoding/json"
	"io/ioutil"
	"time"

	"github.com/gansoi/gansoi/checks"
	"github.com/gansoi/gansoi/database"
	"github.com/gansoi/gansoi/logger"
	"github.com/gansoi/gansoi/notify"
	"github.com/gansoi/gansoi/transports/ssh"
	"github.com/ghodss/yaml"
)

type (
	// Configuration keeps configuration for a core node.
	Configuration struct {
		Bind             string                          `json:"bind"`
		DataDir          string                          `json:"datadir"`
		HTTP             HTTP                            `json:"http"`
		HTTPRedirect     HTTPRedirect                    `json:"redirect"`
		ExclusiveSeeding bool                            `json:"exclusive-seeding"`
		Hosts            map[string]ssh.SSH              `json:"hosts"`
		Checks           map[string]*checks.Check        `json:"checks"`
		ContactGroups    map[string]*notify.ContactGroup `json:"contactgroups"`
		Contacts         map[string]*notify.Contact      `json:"contacts"`

		knownChecks        map[string]bool
		knownHosts         map[string]bool
		knownContactGroups map[string]bool
		knownContacts      map[string]bool
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

// NewConfiguration returns a new configuration with sane defaults.
func NewConfiguration() *Configuration {
	c := &Configuration{
		knownChecks:        make(map[string]bool),
		knownHosts:         make(map[string]bool),
		knownContactGroups: make(map[string]bool),
		knownContacts:      make(map[string]bool),
	}

	// By default we bind to port 443 (HTTPS) on all interfaces on both IPv4
	// and IPv6.
	c.HTTP.Bind = ":443"
	c.HTTP.TLS = true

	c.HTTPRedirect.Bind = ":80"

	c.Bind = ":4934"

	// This makes sense on a unix system.
	c.DataDir = "/var/lib/gansoi"

	return c
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

		c.knownChecks[check.ID] = true
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

func (c *Configuration) SaveContactGroups(w database.Writer) error {
	for id, g := range c.ContactGroups {
		if g.ID == "" {
			g.ID = id
		}

		if g.Name == "" {
			g.Name = id
		}

		err := w.Save(g)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Configuration) SaveContacts(w database.Writer) error {
	for id, contact := range c.Contacts {
		if contact.ID == "" {
			contact.ID = id
		}

		if contact.Name == "" {
			contact.Name = id
		}

		err := w.Save(contact)
		if err != nil {
			return err
		}
	}

	return nil
}

// DeleteUnknownSeeds deletes all checks, hosts, contacts and contactgroups not
// seeded from the configuration.
func (c *Configuration) DeleteUnknownSeeds(db database.ReadWriter) {
	var checks []*checks.Check
	db.All(&checks, -1, 0, false)
	for _, check := range checks {
		if !c.knownChecks[check.ID] {
			logger.Debug("configuration", "Deleting unknown check '%s'", check.ID)
			db.Delete(check)
		}
	}

	var hosts []*ssh.SSH
	db.All(&hosts, -1, 0, false)
	for _, host := range hosts {
		if !c.knownHosts[host.ID] {
			logger.Debug("configuration", "Deleting unknown host '%s'", host.ID)
			db.Delete(host)
		}
	}

	var contactgroups []*notify.ContactGroup
	db.All(&contactgroups, -1, 0, false)
	for _, group := range contactgroups {
		if !c.knownContactGroups[group.ID] {
			logger.Debug("configuration", "Deleting unknown contactgroup '%s'", group.ID)
			db.Delete(group)
		}
	}

	var contacts []*notify.Contact
	db.All(&contacts, -1, 0, false)
	for _, contact := range contacts {
		if !c.knownContacts[contact.ID] {
			logger.Debug("configuration", "Deleting unknown contact '%s'", contact.ID)
			db.Delete(contact)
		}
	}
}
