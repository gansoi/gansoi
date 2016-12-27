package eval

import (
	"time"

	"github.com/gansoi/gansoi/database"
)

type (
	// Evaluation describes the current state of a check when taking all
	// cluster nodes into consideration.
	Evaluation struct {
		CheckID string `storm:"id"`
		Start   time.Time
		End     time.Time
		History States
	}
)

func init() {
	database.RegisterType(Evaluation{})
}
