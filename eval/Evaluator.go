package eval

import (
	"time"

	"github.com/gansoi/gansoi/checks"
	"github.com/gansoi/gansoi/database"
	"github.com/gansoi/gansoi/logger"
)

type (
	// Evaluator will evaluate check results from all nodes on the leader node.
	Evaluator struct {
		db            database.Database
		historyLength int
	}
)

// NewEvaluator will instantiate a new Evaluator listening to cluster changes,
// and evaluating results as they arrive.
func NewEvaluator(db database.Database) *Evaluator {
	e := &Evaluator{
		db:            db,
		historyLength: 5,
	}

	return e
}

func statesFromHistory(history []checks.CheckResult) States {
	var states States
	for _, result := range history {
		state := StateDown

		if result.Error == "" {
			state = StateUp
		}

		states = append(states, state)
	}

	return states
}

// evaluate will FIXME
func (e *Evaluator) evaluate(checkResult *checks.CheckResult) (*Evaluation, error) {
	clock := time.Now()

	// Get latest evaluation.
	eval, _ := LatestEvaluation(e.db, checkResult)
	if eval == nil {
		eval = NewEvaluation(clock, checkResult)
	}
	eval.End = clock

	// Get historyLength checkResults.
	var history []checks.CheckResult
	e.db.Find("CheckHostID", checkResult.CheckHostID, &history, e.historyLength, 0, true)

	if len(history) < e.historyLength {
		logger.Debug("evaluator", "Not enough history for %s yet", checkResult.CheckHostID)
	}

	eval.History = statesFromHistory(history)

	state := StateUnknown
	if len(history) == e.historyLength {
		state = eval.History.Reduce()
	}

	// If the state has changed, we allocate a new evaluation and end the old.
	if eval.State != state {
		e.db.Save(eval)

		nextEval := NewEvaluation(clock, checkResult)
		nextEval.State = state
		nextEval.History = eval.History

		eval = nextEval
	}

	return eval, e.db.Save(eval)
}

// PostApply implements databse.Listener.
func (e *Evaluator) PostApply(leader bool, command database.Command, data interface{}, err error) {
	// If we're not the leader, we abort. Only the leader should evaluate
	// check results.
	if !leader {
		return
	}

	// We're only interested in saves for now.
	if command != database.CommandSave {
		return
	}

	switch data.(type) {
	case *checks.CheckResult:
		e.evaluate(data.(*checks.CheckResult))
	case *Evaluation:
		eval := data.(*Evaluation)
		logger.Debug("eval", "%s: %s (%s) %v", eval.CheckHostID, eval.History.Reduce().ColorString(), eval.End.Sub(eval.Start).String(), eval.History)
	}
}
