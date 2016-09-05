package eval

import (
	"time"

	"github.com/abrander/gansoi/database"
)

type (
	// Evaluation describes the current state of a check when taking all
	// cluster nodes into consideration.
	Evaluation struct {
		CheckID string `storm:"id"`
		State   State
		Start   time.Time
		End     time.Time
	}
)

func init() {
	database.RegisterType(Evaluation{})
}
