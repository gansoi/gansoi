package http

import (
	"testing"

	"github.com/gansoi/gansoi/plugins"
)

func TestAgent(t *testing.T) {
	a := plugins.GetAgent("http")
	_ = a.(*HTTP)
}

var _ plugins.Agent = (*HTTP)(nil)
