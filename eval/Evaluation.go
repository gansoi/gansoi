package eval

import (
	"time"

	"github.com/gansoi/gansoi/database"
)

type (
	// Evaluation describes the current state of a check.
	Evaluation struct {
		ID      uint64 `json:"id" storm:"id"`
		CheckID string `storm:"index"`
		History States
		State   State
		Start   time.Time
		End     time.Time
	}
)

func init() {
	database.RegisterType(Evaluation{})
}

// LatestEvaluation retrieves the latest evaluation if any.
func LatestEvaluation(db database.Database, checkID string) (*Evaluation, error) {
	if checkID == "" {
		return nil, database.ErrNotFound
	}

	var results []Evaluation
	db.Find("CheckID", checkID, &results, 1, 0, true)
	if len(results) != 1 {
		return nil, database.ErrNotFound
	}

	return &results[0], nil
}
