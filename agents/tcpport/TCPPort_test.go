package tcpport

import (
	"testing"

	"github.com/abrander/gansoi/agents"
)

func TestAgent(t *testing.T) {
	a := agents.GetAgent("tcpport")
	_ = a.(*TCPPort)
}

var _ agents.Agent = (*TCPPort)(nil)
