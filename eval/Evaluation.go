package eval

import (
	"time"

	"github.com/gansoi/gansoi/checks"
	"github.com/gansoi/gansoi/database"
)

type (
	// Evaluation describes the current state of a check.
	Evaluation struct {
		ID          uint64    `json:"id" storm:"id"`
		CheckHostID string    `json:"check_host_id" storm:"index"`
		CheckID     string    `json:"check_id" storm:"index"`
		HostID      string    `json:"host_id"`
		History     States    `json:"history"`
		State       State     `json:"state"`
		Start       time.Time `json:"start"`
		End         time.Time `json:"end"`
	}
)

func init() {
	database.RegisterType(Evaluation{})
}

// NewEvaluation returns a new evaluation.
func NewEvaluation(clock time.Time, result *checks.CheckResult) *Evaluation {
	return &Evaluation{
		CheckHostID: result.CheckHostID,
		CheckID:     result.CheckID,
		HostID:      result.HostID,
		Start:       clock,
		End:         clock,
	}
}

// LatestEvaluation retrieves the latest evaluation if any.
func LatestEvaluation(db database.Reader, result *checks.CheckResult) (*Evaluation, error) {
	if result.CheckHostID == "" {
		return nil, database.ErrNotFound
	}

	var results []Evaluation
	db.Find("CheckHostID", result.CheckHostID, &results, 1, 0, true)
	if len(results) != 1 {
		return nil, database.ErrNotFound
	}

	return &results[0], database.ErrNotFound
}
