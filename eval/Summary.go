package eval

import (
	"sync"

	"github.com/gansoi/gansoi/checks"
	"github.com/gansoi/gansoi/database"
)

type (
	// Summary is a summary of all checks.
	Summary struct {
		sync.RWMutex
		ID     string `storm:"id"`
		db     database.Writer
		checks map[string]State
		Checks int           `json:"checks"`
		States map[State]int `json:"states"`
	}
)

func init() {
	database.RegisterType(Summary{})
}

// NewSummary will return a new Summary.
func NewSummary(db database.Writer) *Summary {
	return &Summary{
		ID:     "summary",
		db:     db,
		checks: make(map[string]State),
	}
}

// AddEvaluation will add a new evaluation to the summary.
func (s *Summary) AddEvaluation(eval *Evaluation) error {
	if !eval.Final() {
		return nil
	}

	s.checks[eval.CheckID] = eval.State

	return s.db.Save(s)
}

// PostApply implements database.Listener. We listen for added and deleted checks.
func (s *Summary) PostApply(leader bool, command database.Command, data interface{}) {
	if !leader {
		return
	}

	check, isCheck := data.(*checks.Check)
	evaluation, isEvaluation := data.(*Evaluation)

	s.Lock()
	defer s.Unlock()

	switch {
	case isCheck && command == database.CommandSave:
		s.AddCheck(check.ID)

	case isCheck && command == database.CommandDelete:
		s.RemoveCheck(check.ID)

	case isEvaluation && command == database.CommandSave && evaluation.Final():
		s.AddEvaluation(evaluation)

	default:
		return
	}

	s.summarize()
	_ = s.db.Save(s)
}

// RemoveCheck will remove the check identifier by id from summary.
func (s *Summary) RemoveCheck(id string) {
	delete(s.checks, id)
}

// AddCheck adds a check to the internal list of check IDs.
func (s *Summary) AddCheck(id string) {
	state := s.checks[id]
	s.checks[id] = state
}

// summarize will count and summarize all known and unknown check
func (s *Summary) summarize() {
	s.Checks = 0
	s.States = make(map[State]int)
	s.States[StateUnknown] = 0
	s.States[StateUp] = 0
	s.States[StateDown] = 0

	for _, state := range s.checks {
		s.Checks++
		s.States[state]++
	}
}
