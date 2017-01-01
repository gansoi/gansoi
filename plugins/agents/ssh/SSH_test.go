package ssh

import (
	"testing"

	"github.com/gansoi/gansoi/plugins"
)

func TestAgent(t *testing.T) {
	a := plugins.GetAgent("ssh")
	_ = a.(*SSH)
}

var _ plugins.Agent = (*SSH)(nil)
