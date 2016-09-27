package eval

import (
	"time"

	"github.com/abrander/gansoi/checks"
	"github.com/abrander/gansoi/database"
	"github.com/abrander/gansoi/logger"
	"github.com/abrander/gansoi/node"
	"github.com/hashicorp/raft"
)

type (
	// Evaluator will evaluate check results from all nodes on the leader node.
	Evaluator struct {
		node  *node.Node
		peers raft.PeerStore
	}
)

// NewEvaluator will instantiate a new Evaluator listening to cluster changes,
// and evaluating results as they arrive.
func NewEvaluator(n *node.Node, peers raft.PeerStore) *Evaluator {
	e := &Evaluator{
		node:  n,
		peers: peers,
	}

	n.RegisterListener(e)

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

	err := e.node.One("ID", pe.ID, &pe)
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

	err = e.node.Save(&pe)

	return err
}

// evaluate2 will evalute if a given check should be considered up or down when
// evaluating the result from all nodes.
// This should only be done on the leader.
func (e *Evaluator) evaluate2(n *PartialEvaluation) error {
	var eval Evaluation

	// FIXME: Locking.

	err := e.node.One("CheckID", n.CheckID, &eval)
	if err != nil {
		// Evaluation is unknown. Start new.
		eval.CheckID = n.CheckID
		eval.Start = n.Start
		eval.End = n.End
	}

	state := StateUp
	states := make(map[State]int)

	nodes, _ := e.peers.Peers()
	for _, nodeID := range nodes {
		ID := n.CheckID + ":::" + nodeID

		var pe PartialEvaluation
		err = e.node.One("ID", ID, &pe)
		if err != nil {
			// Not all nodes have reported yet. Could be StateDegraded instead?
			state = StateUnknown
			break
		}

		// FIXME: Deal with old checks somehow. Maybe we should simply discard
		//        them? I'm not sure it makes much sense to evalute a period
		//        where gansoi were not running for example.

		states[pe.State]++
	}

	if state == StateUp {
		if states[StateUp] == len(nodes) {
			state = StateUp
		} else if states[StateDown] == len(nodes) {
			state = StateDown
		} else {
			state = StateDegraded
		}
	}

	if eval.State == state {
		// There is no change in state. Keep current start time, and update end
		// time.
		eval.End = time.Now()
	} else {
		// We have a new state. Update both start and end time.
		eval.Start = time.Now()
		eval.End = eval.Start
	}

	eval.State = state

	return e.node.Save(&eval)
}

// PostClusterApply implements node.Listener.
func (e *Evaluator) PostClusterApply(leader bool, command database.Command, data interface{}, err error) {
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
		logger.Green("eval", "%s: %s (%s)", eval.CheckID, eval.State.ColorString(), eval.End.Sub(eval.Start).String())
	}
}
