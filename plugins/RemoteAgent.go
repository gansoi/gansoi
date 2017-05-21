package plugins

import (
	"reflect"

	"github.com/gansoi/gansoi/transports"
)

type (
	// RemoteAgent should be implemented for all agents supporting a transport.
	RemoteAgent interface {
		RemoteCheck(transport transports.Transport, result AgentResult) error
	}
)

// GetRemoteAgent returns an agent suitable for using a transport.
func GetRemoteAgent(name string) RemoteAgent {
	agent, found := agents[name]
	if !found {
		return nil
	}

	return reflect.New(agent).Interface().(RemoteAgent)
}
