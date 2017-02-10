package eval

import (
	"encoding/json"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/gansoi/gansoi/boltdb"
	"github.com/gansoi/gansoi/checks"
	"github.com/gansoi/gansoi/database"
	"github.com/gansoi/gansoi/plugins"
)

type (
	mockAgent struct {
		ReturnError bool `json:"return_error"`
	}

	peerStore struct {
		peers []string
	}
)

var (
	check = checks.Check{
		AgentID:   "mock",
		Interval:  time.Second,
		Arguments: json.RawMessage("{}"),
	}
)

func (m *mockAgent) Check(result plugins.AgentResult) error {
	if m.ReturnError {
		return errors.New("error")
	}

	result.AddValue("ran", true)

	return nil
}

func init() {
	plugins.RegisterAgent("mock", mockAgent{})

	check.ID = "test"
}

func (p *peerStore) SetPeers(peers []string) error {
	p.peers = peers

	return nil
}

func (p *peerStore) Peers() ([]string, error) {
	return p.peers, nil
}

func newE(t *testing.T, nodes []string) (*boltdb.BoltStore, *Evaluator) {
	path := "/dev/shm/mockdb"
	peers := &peerStore{}
	peers.SetPeers(nodes)

	db, err := boltdb.NewBoltStore(path)
	if err != nil {
		t.Fatalf("Failed to create mock database: %s", err.Error())
	}

	err = os.Remove(path)
	if err != nil {
		t.Fatalf("os.remove() failed: %s", err.Error())
	}

	e := NewEvaluator(db, peers)
	if e == nil {
		t.Fatalf("NewEvaluator() returned nil")
	}

	return db, e
}

func TestEvaluatorEvaluate1Basics(t *testing.T) {
	db, e := newE(t, []string{"justone"})
	defer db.Close()

	result := &checks.CheckResult{
		TimeStamp: time.Now(),
		CheckID:   "test",
		Node:      "justone",
	}

	_, err := e.evaluate1(result)
	if err != nil {
		t.Fatalf("evaluate1() failed: %s", err.Error())
	}

	pe := []PartialEvaluation{}
	err = db.All(&pe, -1, 0, false)
	if err != nil {
		t.Fatalf("db.All() failed: %s", err.Error())
	}

	if len(pe) != 1 {
		t.Fatalf("Got wrong number of evaluations, got %d (%v)", len(pe), pe)
	}

	// Move one minute into the future.
	result.TimeStamp = result.TimeStamp.Add(time.Minute)

	_, err = e.evaluate1(result)
	if err != nil {
		t.Fatalf("evaluate1() failed: %s", err.Error())
	}

	err = db.All(&pe, -1, 0, false)
	if err != nil {
		t.Fatalf("db.All() failed: %s", err.Error())
	}

	if len(pe) != 1 {
		t.Fatalf("Got wrong number of evaluations.")
	}
}

func TestEvaluatorEvaluate1(t *testing.T) {
	db, e := newE(t, []string{"justone"})
	defer db.Close()

	cases := []struct {
		in    checks.CheckResult
		state State
	}{
		{checks.CheckResult{}, StateUp},
		{checks.CheckResult{}, StateUp},
		{checks.CheckResult{Error: "error"}, StateDown},
		{checks.CheckResult{Error: "error"}, StateDown},
		{checks.CheckResult{Error: "error"}, StateDown},
		{checks.CheckResult{}, StateUp},
		{checks.CheckResult{Error: "error"}, StateDown},
		{checks.CheckResult{}, StateUp},
	}

	for _, c := range cases {
		e, _ := e.evaluate1(&c.in)
		if e.State != c.state {
			t.Fatalf("evaluate1() concluded wrong state. Got %s, expected %s", e.State.ColorString(), c.state.ColorString())
		}
	}
}

func TestEvaluatorEvaluate2(t *testing.T) {
	db, e := newE(t, []string{"justone"})
	defer db.Close()

	err := db.Save(&check)
	if err != nil {
		t.Fatalf("db.Save() failed: %s", err.Error())
	}

	cases := []struct {
		in    checks.CheckResult
		state State
	}{
		{checks.CheckResult{CheckID: "test", Node: "justone", TimeStamp: time.Now()}, StateUp},
		{checks.CheckResult{CheckID: "test", Node: "justone", TimeStamp: time.Now()}, StateUp},
		{checks.CheckResult{CheckID: "test", Node: "justone", TimeStamp: time.Now(), Error: "error"}, StateDown},
		{checks.CheckResult{CheckID: "test", Node: "justone", TimeStamp: time.Now(), Error: "error"}, StateDown},
		{checks.CheckResult{CheckID: "test", Node: "justone", TimeStamp: time.Now(), Error: "error"}, StateDown},
		{checks.CheckResult{CheckID: "test", Node: "justone", TimeStamp: time.Now(), Error: "error"}, StateDown},
		{checks.CheckResult{CheckID: "test", Node: "justone", TimeStamp: time.Now()}, StateUp},
		{checks.CheckResult{CheckID: "test", Node: "justone", TimeStamp: time.Now()}, StateUp},
		{checks.CheckResult{CheckID: "test", Node: "justone", TimeStamp: time.Now()}, StateUp},
		{checks.CheckResult{CheckID: "test", Node: "justone", TimeStamp: time.Now()}, StateUp},
		{checks.CheckResult{CheckID: "test", Node: "justone", TimeStamp: time.Now()}, StateUp},
		{checks.CheckResult{CheckID: "test", Node: "justone", TimeStamp: time.Now()}, StateUp},
		{checks.CheckResult{CheckID: "test", Node: "justone", TimeStamp: time.Now()}, StateUp},
		{checks.CheckResult{CheckID: "test", Node: "justone", TimeStamp: time.Now()}, StateUp},
		{checks.CheckResult{CheckID: "test", Node: "justone", TimeStamp: time.Now()}, StateUp},
	}

	for i, c := range cases {
		pe, _ := e.evaluate1(&c.in)
		eval, err := e.evaluate2(pe)
		if err != nil {
			t.Fatalf("evaluate2() returned error: %s", err.Error())
		}

		evalState := eval.History[0]
		if evalState != c.state {
			t.Fatalf("evaluate2(%d) concluded wrong state. Got %s, expected %s", i, evalState.ColorString(), c.state.ColorString())
		}
	}
}

func TestEvaluatorEvaluate2Cluster(t *testing.T) {
	db, e := newE(t, []string{"one", "two"})
	defer db.Close()

	err := db.Save(&check)
	if err != nil {
		t.Fatalf("db.Save() failed: %s", err.Error())
	}

	cases := []struct {
		in    checks.CheckResult
		state State
	}{
		{checks.CheckResult{CheckID: "test", Node: "one", TimeStamp: time.Now()}, StateUnknown},
		{checks.CheckResult{CheckID: "test", Node: "one", TimeStamp: time.Now()}, StateUnknown},
		{checks.CheckResult{CheckID: "test", Node: "two", TimeStamp: time.Now()}, StateUp},
		{checks.CheckResult{CheckID: "test", Node: "two", TimeStamp: time.Now()}, StateUp},
		{checks.CheckResult{CheckID: "test", Node: "one", TimeStamp: time.Now(), Error: "error"}, StateDegraded},
		{checks.CheckResult{CheckID: "test", Node: "two", TimeStamp: time.Now(), Error: "error"}, StateDown},
		{checks.CheckResult{CheckID: "test", Node: "one", TimeStamp: time.Now(), Error: "error"}, StateDown},
		{checks.CheckResult{CheckID: "test", Node: "one", TimeStamp: time.Now(), Error: "error"}, StateDown},
		{checks.CheckResult{CheckID: "test", Node: "one", TimeStamp: time.Now()}, StateDegraded},
		{checks.CheckResult{CheckID: "test", Node: "two", TimeStamp: time.Now()}, StateUp},
		{checks.CheckResult{CheckID: "test", Node: "one", TimeStamp: time.Now()}, StateUp},
		{checks.CheckResult{CheckID: "test", Node: "one", TimeStamp: time.Now()}, StateUp},
		{checks.CheckResult{CheckID: "test", Node: "one", TimeStamp: time.Now(), Error: "error"}, StateDegraded},
		{checks.CheckResult{CheckID: "test", Node: "one", TimeStamp: time.Now()}, StateUp},
		{checks.CheckResult{CheckID: "test", Node: "one", TimeStamp: time.Now()}, StateUp},

		// Old results should result in unknown. Something is wrong.
		{checks.CheckResult{CheckID: "test", Node: "one", TimeStamp: time.Now().Add(-time.Hour)}, StateUnknown},
		{checks.CheckResult{CheckID: "test", Node: "two", TimeStamp: time.Now().Add(-time.Hour)}, StateUnknown},
		{checks.CheckResult{CheckID: "test", Node: "one", TimeStamp: time.Now()}, StateUnknown},
		{checks.CheckResult{CheckID: "test", Node: "one", TimeStamp: time.Now()}, StateUnknown},
		{checks.CheckResult{CheckID: "test", Node: "two", TimeStamp: time.Now()}, StateUp},
	}

	for i, c := range cases {
		pe, _ := e.evaluate1(&c.in)
		eval, err := e.evaluate2(pe)
		if err != nil {
			t.Fatalf("evaluate2() returned error: %s", err.Error())
		}

		evalState := eval.History[0]
		if evalState != c.state {
			t.Fatalf("evaluate2(%d) concluded wrong state. Got %s, expected %s", i, evalState.ColorString(), c.state.ColorString())
		}
	}
}

var _ database.Listener = (*Evaluator)(nil)
