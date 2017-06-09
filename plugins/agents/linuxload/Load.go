package load

import (
	"errors"
	"fmt"

	"github.com/gansoi/gansoi/plugins"
	"github.com/gansoi/gansoi/transports"
)

type (
	// Load reads load and process data from Linux hosts.
	Load struct{}
)

var (
	// ErrSyntax will be returned, if we don't understand the format of /proc/loadavg.
	ErrSyntax = errors.New("Unknown format of /proc/loadavg")
)

func init() {
	plugins.RegisterAgent("linuxload", Load{})
}

// RemoteCheck implements RemoteAgent.
func (l *Load) RemoteCheck(transport transports.Transport, result plugins.AgentResult) error {
	contents, err := transport.ReadFile("/proc/loadavg")
	if err != nil {
		return err
	}

	var load1 float32
	var load5 float32
	var load15 float32
	var running int
	var total int
	var lastPid int

	_, err = fmt.Sscanf(string(contents), "%f %f %f %d/%d %d", &load1, &load5, &load15, &running, &total, &lastPid)

	if err != nil {
		return ErrSyntax
	}

	result.AddValue("Load1", load1)
	result.AddValue("Load5", load5)
	result.AddValue("Load15", load15)
	result.AddValue("Running", running)
	result.AddValue("Processes", total)

	return nil
}
