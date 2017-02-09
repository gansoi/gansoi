package cluster

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"strings"
	"sync"
)

type (
	// Info stores the current list of peers.
	Info struct {
		sync.RWMutex
		path         string
		SelfName     string   `json:"self"`
		PeerList     []string `json:"peers"`
		CACert       []byte   `json:"ca-cert"`
		CAKey        []byte   `json:"ca-key"`
		NodeCert     []byte   `json:"node-cert"`
		NodeKey      []byte   `json:"node-key"`
		ClusterToken string   `json:"cluster-token"`
	}
)

// NewInfo will instanmtiate a new peer store that will save its state in
// path.
func NewInfo(path string) *Info {
	c := &Info{
		path: path,
	}

	c.Load()

	return c
}

// Save will trigger a save.
func (c *Info) Save() error {
	c.RLock()
	b, err := json.MarshalIndent(c, "", "\t")
	c.RUnlock()
	if err != nil {
		return err
	}

	return ioutil.WriteFile(c.path, b, 0600)
}

// Load loads the state from persistent disk storage.
func (c *Info) Load() error {
	b, err := ioutil.ReadFile(c.path)
	if err != nil {
		return err
	}

	c.Lock()
	err = json.Unmarshal(b, c)
	c.Unlock()

	return err
}

// Peers returns the list of known peers.
func (c *Info) Peers() ([]string, error) {
	c.RLock()
	peers := c.PeerList
	c.RUnlock()

	return peers, nil
}

// SetPeers updates the list of peers.
func (c *Info) SetPeers(peers []string) error {
	c.Lock()
	c.PeerList = peers
	c.Unlock()

	return c.Save()
}

// Self will return our own name as set by SetSelf().
func (c *Info) Self() string {
	c.RLock()
	self := c.SelfName
	c.RUnlock()

	return self
}

// IP will return our presumed cluster IP. If we have no way of guessing it,
// IP() simply returns nil.
func (c *Info) IP() net.IP {
	self := c.Self()

	s, err := net.ResolveTCPAddr("tcp", DefaultPort(self))
	if err != nil {
		return nil
	}

	return s.IP
}

// SetSelf will set our own node name.
func (c *Info) SetSelf(self string) error {
	c.Lock()
	c.SelfName = self
	c.Unlock()

	return c.Save()
}

// DefaultPort will return the hostname with the default internal Gansoi port.
// If hostport already contains a port, DefaultPort will simply return that.
func DefaultPort(hostport string) string {
	if strings.IndexRune(hostport, ':') < 0 {
		return hostport + ":" + "4934"
	}

	return hostport
}
