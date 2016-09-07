package http

import (
	"testing"

	"github.com/abrander/gansoi/agents"
)

func TestAgent(t *testing.T) {
	a := agents.GetAgent("http")
	_ = a.(*HTTP)
}

var _ agents.Agent = (*HTTP)(nil)
