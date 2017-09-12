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
		db            database.ReadWriter
		historyLength int
	}
)

// NewEvaluator will instantiate a new Evaluator listening to cluster changes,
// and evaluating results as they arrive.
func NewEvaluator(db database.ReadWriter) *Evaluator {
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
func (e *Evaluator) Evaluate(checkResult *checks.CheckResult) (*Evaluation, error) {
	clock := time.Now()

	// Get latest evaluation.
	eval, _ := LatestEvaluation(e.db, checkResult)
	if eval == nil {
		eval = NewEvaluation(clock, checkResult)
	}
	eval.End = clock

	if len(eval.Results) >= e.historyLength {
		eval.Results = eval.Results[len(eval.Results)-e.historyLength+1 : 5]
	}

	results := append(eval.Results, *checkResult)
	history := statesFromHistory(results)

	state := StateUnknown
	if len(results) == e.historyLength {
		state = history.Reduce()
	}

	// If the state has changed, we allocate a new evaluation and end the old.
	if eval.State != state {
		eval.Save(e.db)

		nextEval := NewEvaluation(clock, checkResult)
		nextEval.State = state

		eval = nextEval
	}

	eval.History = history
	eval.Results = results

	logger.Debug("eval", "%s: %s (%s) %s", eval.CheckHostID, eval.History.Reduce().ColorString(), eval.End.Sub(eval.Start).String(), eval.History.ColorString())

	if eval.HostID != "" {
		e.evaluteHost(eval)
	}

	return eval, eval.Save(e.db)
}

func (e *Evaluator) evaluteHost(hostEval *Evaluation) (*Evaluation, error) {
	result := &checks.CheckResult{
		CheckID:     hostEval.CheckID,
		CheckHostID: checks.CheckHostID(hostEval.CheckID, ""),
	}

	eval, _ := LatestEvaluation(e.db, result)
	if eval == nil {
		eval = NewEvaluation(hostEval.Start, result)
		eval.History = States{StateUnknown}
	}
	eval.End = hostEval.End

	var check checks.Check
	err := e.db.One("ID", eval.CheckID, &check)
	if err != nil {
		return nil, err
	}

	eval.Hosts[hostEval.HostID] = hostEval.State

	var state State
	states := make(map[State]int)
	for _, key := range check.Hosts {
		states[eval.Hosts[key]]++
	}

	switch true {
	case states[StateUnknown] > 0:
		state = StateUnknown
	case states[StateDown] > 0:
		state = StateDown
	case states[StateUp] == len(check.Hosts):
		state = StateUp
	}

	// If the state has changed, we allocate a new evaluation and end the old.
	if eval.State != state {
		eval.Save(e.db)

		nextEval := NewEvaluation(hostEval.End, result)
		nextEval.State = state
		nextEval.History = States{state}
		nextEval.Hosts = eval.Hosts

		eval = nextEval
	}

	return eval, eval.Save(e.db)
}
