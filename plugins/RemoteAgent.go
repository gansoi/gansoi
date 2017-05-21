package plugins

import (
	"github.com/gansoi/gansoi/transports"
)

type (
	// RemoteAgent should be implemented for all agents supporting a transport.
	RemoteAgent interface {
		RemoteCheck(transport transports.Transport, result AgentResult) error
	}
)
