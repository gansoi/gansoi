package eval

import (
	"sync"
	"time"

	"github.com/gansoi/gansoi/checks"
	"github.com/gansoi/gansoi/database"
)

type (
	// Evaluation describes the current state of a check.
	Evaluation struct {
		ID          int64                `json:"id" storm:"id,increment"`
		CheckHostID string               `json:"check_host_id" storm:"index"`
		CheckID     string               `json:"check_id" storm:"index"`
		HostID      string               `json:"host_id"`
		History     States               `json:"history"`
		State       State                `json:"state"`
		Start       time.Time            `json:"start"`
		End         time.Time            `json:"end"`
		Hosts       map[string]State     `json:"hosts"`
		Results     []checks.CheckResult `json:"-"`
	}
)

var (
	cacheLock sync.RWMutex
	cache     = make(map[string]*Evaluation)
)

func init() {
	database.RegisterType(Evaluation{})
}

// NewEvaluation returns a new evaluation.
func NewEvaluation(clock time.Time, result *checks.CheckResult) *Evaluation {
	return &Evaluation{
		ID:          time.Now().UnixNano(),
		CheckHostID: result.CheckHostID,
		CheckID:     result.CheckID,
		HostID:      result.HostID,
		Start:       clock,
		End:         clock,
		Hosts:       make(map[string]State),
	}
}

// Save will save e to the supplied db and cache.
func (e *Evaluation) Save(db database.Writer) error {
	err := db.Save(e)

	if err == nil {
		cacheLock.Lock()
		cache[e.CheckHostID] = e
		cacheLock.Unlock()
	}

	return err
}

// LatestEvaluation retrieves the latest evaluation if any.
func LatestEvaluation(db database.Reader, result *checks.CheckResult) (*Evaluation, error) {
	if result.CheckHostID == "" {
		return nil, database.ErrNotFound
	}

	cacheLock.RLock()
	eval, found := cache[result.CheckHostID]
	cacheLock.RUnlock()

	if found {
		return eval, nil
	}

	var results []Evaluation
	db.Find("CheckHostID", result.CheckHostID, &results, 1, 0, true)
	if len(results) != 1 {
		return nil, database.ErrNotFound
	}

	return &results[0], nil
}

// Final returns true if the evaluation is for a check and not for a host-part
// of a check.
func (e *Evaluation) Final() bool {
	return e.CheckHostID == checks.CheckHostID(e.CheckID, "")
}
