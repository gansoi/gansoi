package cluster

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"os"
	"path"
	"strings"
	"sync"
	"syscall"
)

type (
	// Info stores the current list of peers.
	Info struct {
		sync.RWMutex
		path         string
		SelfName     string `json:"self"`
		CACert       []byte `json:"ca-cert"`
		CAKey        []byte `json:"ca-key"`
		NodeCert     []byte `json:"node-cert"`
		NodeKey      []byte `json:"node-key"`
		ClusterToken string `json:"cluster-token"`
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
	b, _ := json.MarshalIndent(c, "", "\t")
	c.RUnlock()

	// We try to make the file owner by the directory owner, if the file
	// doesn't exist.
	if _, err := os.Stat(c.path); os.IsNotExist(err) {
		var st syscall.Stat_t

		// If the parent directory doesn't exist, we should return an error.
		err = syscall.Stat(path.Dir(c.path), &st)
		if err != nil {
			return err
		}

		// We do this without any form of error handling, if it fails, it
		// fails. Users can have reason for wanting this.
		ioutil.WriteFile(c.path, []byte(""), 0600)
		os.Chown(c.path, int(st.Uid), int(st.Gid))
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
	if !strings.ContainsRune(hostport, ':') {
		return hostport + ":" + "4934"
	}

	return hostport
}
