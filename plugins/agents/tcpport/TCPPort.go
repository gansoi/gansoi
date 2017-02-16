package tcpport

import (
	"net"
	"time"

	"github.com/gansoi/gansoi/plugins"
)

func init() {
	plugins.RegisterAgent("tcpport", TCPPort{})
}

// TCPPort will connect to a tcp port and measure timing.
type TCPPort struct {
	Address string `json:"address" description:"The address to connect to (host:port)"`
}

// Check implements plugins.Agent.
func (t *TCPPort) Check(result plugins.AgentResult) error {
	start := time.Now()
	conn, err := net.Dial("tcp", t.Address)
	if err != nil {
		return err
	}

	// Measure the duration. This is the only check we for for now.
	result.AddValue("ConnectDuration", ms(time.Now().Sub(start)))

	// No need to check for errors here, it is not an error if the remote end
	// already closed the connection.
	conn.Close()

	return nil
}

// ms will convert a time.Duration to milliseconds.
func ms(d time.Duration) int64 {
	return ((d + time.Millisecond/2) / time.Millisecond).Nanoseconds()
}
