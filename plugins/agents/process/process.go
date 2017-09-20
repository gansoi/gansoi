package process

import (
	"bytes"
	"io/ioutil"

	"github.com/pkg/errors"

	"github.com/gansoi/gansoi/plugins"
	"github.com/gansoi/gansoi/transports"
)

func init() {
	plugins.RegisterAgent("process", Process{})
}

// Process will check if a process with exact name is running on the host.
type Process struct {
	Name string `json:"name" description:"Exact process name"`
}

// RemoteCheck implements plugins.RemoteAgent.
func (m *Process) RemoteCheck(transport transports.Transport, result plugins.AgentResult) error {
	out, _, err := transport.Exec("pidof", m.Name)
	if err != nil {
		return errors.Wrap(err, "pidof")
	}

	resultAsBytes, err := ioutil.ReadAll(out)
	if err != nil {
		return err
	}

	result.AddValue("Running", len(bytes.Fields(resultAsBytes)))
	return nil
}
