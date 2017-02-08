package eval

import (
	"fmt"
	"time"

	"github.com/hashicorp/raft"

	"github.com/gansoi/gansoi/checks"
	"github.com/gansoi/gansoi/database"
	"github.com/gansoi/gansoi/logger"
)

type (
	// Evaluator will evaluate check results from all nodes on the leader node.
	Evaluator struct {
		db            database.Database
		peers         raft.PeerStore
		historyLength int
	}
)

// NewEvaluator will instantiate a new Evaluator listening to cluster changes,
// and evaluating results as they arrive.
func NewEvaluator(db database.Database, peers raft.PeerStore) *Evaluator {
	e := &Evaluator{
		db:    db,
		peers: peers,
	}

	// Preserve state history for double the cluster size. This ensures that
	// we have at least to results from at least two runs from each node.
	p, _ := peers.Peers()
	e.historyLength = len(p) * 2

	db.RegisterListener(e)

	return e
}

// evaluate1 will evalute a check result from a single node.
func (e *Evaluator) evaluate1(checkResult *checks.CheckResult) error {
	pe := PartialEvaluation{}
	pe.ID = checkResult.CheckID + ":::" + checkResult.Node

	state := StateDown

	// Evaluate if the check went well for a single node. For now we simply
	// checks for empty error string and assume everything is good if its
	// empty :)
	if checkResult.Error == "" {
		state = StateUp
	}

	err := e.db.One("ID", pe.ID, &pe)
	if err != nil {
		// None was found. Fill out new.
		pe.CheckID = checkResult.CheckID
		pe.NodeID = checkResult.Node
		pe.Start = checkResult.TimeStamp
		pe.End = checkResult.TimeStamp
		pe.State = state
	} else {
		// Check if the state changed.
		if pe.State == state {
			// If the state is the same, just update end time.
			pe.End = checkResult.TimeStamp
		} else {
			// If the state changed, reset both start and end time.
			pe.Start = checkResult.TimeStamp
			pe.End = checkResult.TimeStamp
		}

		pe.State = state
	}

	err = e.db.Save(&pe)

	return err
}

// evaluate2 will evalute if a given check should be considered up or down when
// evaluating the result from all nodes.
// This should only be done on the leader.
func (e *Evaluator) evaluate2(n *PartialEvaluation) error {
	var eval Evaluation

	// FIXME: Locking.

	err := e.db.One("CheckID", n.CheckID, &eval)
	if err != nil {
		// Evaluation is unknown. Start new.
		eval.CheckID = n.CheckID
		eval.Start = n.Start
		eval.End = n.End
	}

	check := checks.Check{}
	err = e.db.One("ID", n.CheckID, &check)
	if err != nil {
		return fmt.Errorf("Got result from unknown check [%s]", n.CheckID)
	}

	state := StateUnknown

	nodes, _ := e.peers.Peers()
	nodeStates := States{}

	for _, nodeID := range nodes {
		ID := n.CheckID + ":::" + nodeID

		var pe PartialEvaluation
		err = e.db.One("ID", ID, &pe)
		if err != nil {
			// Not all nodes have reported yet, abort.
			break
		}

		// If any result is older than two cycles, we discard it.
		if time.Now().Sub(pe.End) > check.Interval*2 {
			logger.Debug("eval", "Result from %s is too old (T:%s) (D:%s) (I:%s)", nodeID, pe.End, time.Now().Sub(pe.End), check.Interval)
			break
		}

		nodeStates.Add(pe.State, -1)
	}

	// Only assume anything other than StateUnknown if all nodes have reported.
	if len(nodeStates) == len(nodes) {
		state = nodeStates.Reduce()
	}

	lastReduce := eval.History.Reduce()
	eval.History.Add(state, e.historyLength)

	if eval.History.Reduce() == lastReduce {
		// There is no change in state. Keep current start time, and update end
		// time.
		eval.End = time.Now()
	} else {
		// We have a new state. Update both start and end time.
		eval.Start = time.Now()
		eval.End = eval.Start
	}

	return e.db.Save(&eval)
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
		e.evaluate1(data.(*checks.CheckResult))
	case *PartialEvaluation:
		e.evaluate2(data.(*PartialEvaluation))
	case *Evaluation:
		eval := data.(*Evaluation)
		logger.Debug("eval", "%s: %s (%s) %v", eval.CheckID, eval.History.Reduce().ColorString(), eval.End.Sub(eval.Start).String(), eval.History)
	}
}
