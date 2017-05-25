package eval

import (
	"time"

	"github.com/gansoi/gansoi/checks"
	"github.com/gansoi/gansoi/database"
)

type (
	// Evaluation describes the current state of a check.
	Evaluation struct {
		ID      uint64    `json:"id" storm:"id"`
		CheckID string    `json:"check_id" storm:"index"`
		HostID  string    `json:"host_id" storm:"index"`
		History States    `json:"history"`
		State   State     `json:"state"`
		Start   time.Time `json:"start"`
		End     time.Time `json:"end"`
	}
)

func init() {
	database.RegisterType(Evaluation{})
}

// NewEvaluation returns a new evaluation.
func NewEvaluation(clock time.Time, result *checks.CheckResult) *Evaluation {
	return &Evaluation{
		CheckID: result.CheckID,
		HostID:  result.HostID,
		Start:   clock,
		End:     clock,
	}
}

// LatestEvaluation retrieves the latest evaluation if any.
func LatestEvaluation(db database.Database, result *checks.CheckResult) (*Evaluation, error) {
	if result.CheckID == "" {
		return nil, database.ErrNotFound
	}

	var evals []Evaluation
	// FIXME: Query using checkID *and* hostID.
	db.Find("CheckID", result.CheckID, &evals, -1, 0, true)
	for _, eval := range evals {
		if eval.HostID == result.HostID {
			return &eval, nil
		}
	}

	return nil, database.ErrNotFound
}
