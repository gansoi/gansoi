package http

import (
	"testing"

	"github.com/abrander/gansoi/plugins"
)

func TestAgent(t *testing.T) {
	a := plugins.GetAgent("http")
	_ = a.(*HTTP)
}

var _ plugins.Agent = (*HTTP)(nil)
