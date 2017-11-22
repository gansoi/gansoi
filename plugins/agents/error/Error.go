package error

import (
	"fmt"
	"math/rand"

	"github.com/gansoi/gansoi/plugins"
)

type (
	// Error is only useful for testing Gansoi.
	Error struct {
		Chance int `json:"chance" description:"What's the chance of failure (0-100)" default:"50"`
	}
)

func init() {
	plugins.RegisterAgent("error", Error{})
}

// Check implements plugins.Agent.
func (e *Error) Check(result plugins.AgentResult) error {
	n := rand.Intn(99) + 1

	if n <= e.Chance {
		return fmt.Errorf("%d was below %d", n, e.Chance)
	}

	return nil
}
