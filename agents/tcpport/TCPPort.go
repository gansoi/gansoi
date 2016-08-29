package tcpport

import (
	"net"
	"time"

	"github.com/abrander/gansoi/agents"
)

func init() {
	agents.RegisterAgent("tcpport", TCPPort{})
}

// TCPPort will connect to a tcp port and measure timing.
type TCPPort struct {
	Address string `json:"address" description:"The address to connect to (host:port)"`
}

// Result is the result from this test.
type Result struct {
	ConnectDuration time.Duration `description:"The time it took to connect"`
}

// Check implements agents.Agent.
func (t *TCPPort) Check() (interface{}, error) {
	r := &Result{}

	start := time.Now()
	conn, err := net.Dial("tcp", t.Address)
	if err != nil {
		return nil, err
	}

	// Measure the duration. This is the only check we for for now.
	r.ConnectDuration = time.Now().Sub(start)

	// It doesn't make sense to measure close timing. Go returns without error
	// before the remote end acks.
	err = conn.Close()
	if err != nil {
		return nil, err
	}

	return r, nil
}
