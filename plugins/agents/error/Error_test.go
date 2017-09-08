package error

import (
	"testing"

	"github.com/gansoi/gansoi/plugins"
)

func TestAgent(t *testing.T) {
	a := plugins.GetAgent("error")
	_ = a.(*Error)
}

func TestCheckFail(t *testing.T) {
	a := Error{
		Chance: 100,
	}

	result := plugins.NewAgentResult()
	err := a.Check(result)
	if err == nil {
		t.Fatalf("Failed to error")
	}
}

func TestCheckl(t *testing.T) {
	a := Error{
		Chance: 0,
	}

	result := plugins.NewAgentResult()
	err := a.Check(result)
	if err != nil {
		t.Fatalf("errored")
	}
}
