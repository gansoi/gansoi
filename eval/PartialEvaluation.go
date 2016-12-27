package eval

import (
	"time"

	"github.com/gansoi/gansoi/database"
)

type (
	// PartialEvaluation describes a summary for a check for a specific node.
	PartialEvaluation struct {
		ID      string `storm:"id"` // A composit between checkid and nodeid.
		CheckID string
		NodeID  string
		State   State
		Start   time.Time
		End     time.Time
	}
)

func init() {
	database.RegisterType(PartialEvaluation{})
}
