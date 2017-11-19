package smtp

import (
	"net/textproto"
	"strings"

	"github.com/gansoi/gansoi/plugins"
)

type (
	// SMTP is a simple agent for SMTP servers.
	SMTP struct {
		Address string `json:"address" description:"The address to connect to (host or host:port)"`
	}
)

func init() {
	plugins.RegisterAgent("smtp", SMTP{})
}

// defaultPort will append the default ssh port to a hostname if needed.
func defaultPort(address string) string {
	if !strings.ContainsRune(address, ':') {
		return address + ":25"
	}

	return address
}

// Check implements plugins.Agent.
func (s *SMTP) Check(result plugins.AgentResult) error {
	conn, err := textproto.Dial("tcp", defaultPort(s.Address))
	if err != nil {
		return err
	}
	defer conn.Close()

	_, banner, _ := conn.ReadResponse(220)

	result.AddValue("banner", banner)

	return nil
}
