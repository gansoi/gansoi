package tcpport

import (
	"testing"

	"github.com/abrander/gansoi/plugins"
)

func TestAgent(t *testing.T) {
	a := plugins.GetAgent("tcpport")
	_ = a.(*TCPPort)
}

var _ plugins.Agent = (*TCPPort)(nil)
