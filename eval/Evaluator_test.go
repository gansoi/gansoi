package eval

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/gansoi/gansoi/boltdb"
	"github.com/gansoi/gansoi/checks"
	"github.com/gansoi/gansoi/plugins"
)

type (
	mockAgent struct {
		ReturnError bool `json:"return_error"`
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

func newE(t *testing.T) (*boltdb.TestStore, *Evaluator) {
	db := boltdb.NewTestStore()

	e := NewEvaluator(db)
	if e == nil {
		t.Fatalf("NewEvaluator() returned nil")
	}

	return db, e
}

func TestEvaluatorEvaluate1Basics(t *testing.T) {
	db, e := newE(t)
	defer db.Close()

	result := &checks.CheckResult{
		TimeStamp:   time.Now(),
		CheckHostID: "da::",
		Node:        "justone",
	}

	_, err := e.Evaluate(result)
	if err != nil {
		t.Fatalf("evaluate() failed: %s", err.Error())
	}

	pe := []Evaluation{}
	err = db.All(&pe, -1, 0, false)
	if err != nil {
		t.Fatalf("db.All() failed: %s", err.Error())
	}

	if len(pe) != 1 {
		t.Fatalf("Got wrong number of evaluations, got %d (%v)", len(pe), pe)
	}

	// Move one minute into the future.
	result.TimeStamp = result.TimeStamp.Add(time.Minute)

	_, err = e.Evaluate(result)
	if err != nil {
		t.Fatalf("evaluate() failed: %s", err.Error())
	}

	err = db.All(&pe, -1, 0, false)
	if err != nil {
		t.Fatalf("db.All() failed: %s", err.Error())
	}

	if len(pe) != 1 {
		t.Fatalf("Got wrong number of evaluations, got %d", len(pe))
	}
}

func TestEvaluatorEvaluate(t *testing.T) {
	db, e := newE(t)
	defer db.Close()

	cases := []struct {
		in    checks.CheckResult
		state State
	}{
		{checks.CheckResult{CheckHostID: "aaa"}, StateUnknown},
		{checks.CheckResult{CheckHostID: "aaa"}, StateUnknown},
		{checks.CheckResult{CheckHostID: "aaa"}, StateUnknown},
		{checks.CheckResult{CheckHostID: "aaa"}, StateUnknown},
		{checks.CheckResult{CheckHostID: "aaa"}, StateUp},
		{checks.CheckResult{CheckHostID: "aaa"}, StateUp},
		{checks.CheckResult{CheckHostID: "aaa"}, StateUp},
		{checks.CheckResult{CheckHostID: "aaa", Error: "error"}, StateUp},
		{checks.CheckResult{CheckHostID: "aaa", Error: "error"}, StateUp},
		{checks.CheckResult{CheckHostID: "aaa", Error: "error"}, StateDown},
		{checks.CheckResult{CheckHostID: "aaa"}, StateDown},
		{checks.CheckResult{CheckHostID: "aaa", Error: "error"}, StateDown},
		{checks.CheckResult{CheckHostID: "aaa"}, StateDown},
		{checks.CheckResult{CheckHostID: "aaa"}, StateUp},
		{checks.CheckResult{CheckHostID: "aaa"}, StateUp},
		{checks.CheckResult{CheckHostID: "aaa", Error: "error"}, StateUp},
		{checks.CheckResult{CheckHostID: "aaa"}, StateUp},
		{checks.CheckResult{CheckHostID: "aaa", Error: "error"}, StateUp},
		{checks.CheckResult{CheckHostID: "aaa"}, StateUp},
		{checks.CheckResult{CheckHostID: "aaa"}, StateUp},
		{checks.CheckResult{CheckHostID: "aaa", Error: "error"}, StateUp},
		{checks.CheckResult{CheckHostID: "aaa"}, StateUp},
		{checks.CheckResult{CheckHostID: "aaa"}, StateUp},
		{checks.CheckResult{CheckHostID: "aaa", Error: "error"}, StateUp},
		{checks.CheckResult{CheckHostID: "aaa", Error: "error"}, StateDown},
		{checks.CheckResult{CheckHostID: "aaa", Error: "error"}, StateDown},
		{checks.CheckResult{CheckHostID: "aaa", Error: "error"}, StateDown},
		{checks.CheckResult{CheckHostID: "aaa"}, StateDown},
		{checks.CheckResult{CheckHostID: "aaa"}, StateDown},
		{checks.CheckResult{CheckHostID: "aaa", Error: "error"}, StateDown},
		{checks.CheckResult{CheckHostID: "aaa"}, StateUp},
		{checks.CheckResult{CheckHostID: "aaa"}, StateUp},
	}

	for i, c := range cases {
		err := db.Save(&c.in)
		if err != nil {
			t.Fatalf("Save() failed: %s", err.Error())
		}

		e, _ := e.Evaluate(&c.in)
		if e.State != c.state {
			t.Fatalf("evaluate() [%d] concluded wrong state. Got %s, expected %s", i, e.State.ColorString(), c.state.ColorString())
		}
	}
}

func TestEvaluatorEvaluateHost(t *testing.T) {
	db, e := newE(t)
	defer db.Close()

	c := &checks.Check{
		Hosts: []string{"hostid1", "hostid2"},
	}
	c.ID = "cid1"

	db.Save(c)

	cases := []struct {
		in     checks.CheckResult
		state1 State
		state2 State
	}{
		{checks.CheckResult{CheckID: "unknown-check", CheckHostID: "bbb", HostID: "hostid1"}, StateUnknown, StateUnknown},
		{checks.CheckResult{CheckID: "cid1", CheckHostID: "bbb", HostID: "hostid1"}, StateUnknown, StateUnknown},
		{checks.CheckResult{CheckID: "cid1", CheckHostID: "bbb", HostID: "hostid2"}, StateUnknown, StateUnknown},
		{checks.CheckResult{CheckID: "cid1", CheckHostID: "bbb", HostID: "hostid1"}, StateUnknown, StateUnknown},
		{checks.CheckResult{CheckID: "cid1", CheckHostID: "bbb", HostID: "hostid2"}, StateUp, StateUnknown},
		{checks.CheckResult{CheckID: "cid1", CheckHostID: "bbb", HostID: "hostid1"}, StateUp, StateUnknown},
		{checks.CheckResult{CheckID: "cid1", CheckHostID: "bbb", HostID: "hostid2"}, StateUp, StateUnknown},
		{checks.CheckResult{CheckID: "cid1", CheckHostID: "bbb", HostID: "hostid1", Error: "error"}, StateUp, StateUnknown},
		{checks.CheckResult{CheckID: "cid1", CheckHostID: "bbb", HostID: "hostid2", Error: "error"}, StateUp, StateUnknown},
		{checks.CheckResult{CheckID: "cid1", CheckHostID: "bbb", HostID: "hostid1", Error: "error"}, StateDown, StateDown},
		{checks.CheckResult{CheckID: "cid1", CheckHostID: "bbb", HostID: "hostid2"}, StateDown, StateDown},
		{checks.CheckResult{CheckID: "cid1", CheckHostID: "bbb", HostID: "hostid1", Error: "error"}, StateDown, StateDown},
		{checks.CheckResult{CheckID: "cid1", CheckHostID: "bbb", HostID: "hostid2"}, StateDown, StateDown},
		{checks.CheckResult{CheckID: "cid1", CheckHostID: "bbb", HostID: "hostid1"}, StateUp, StateUp},
		{checks.CheckResult{CheckID: "cid1", CheckHostID: "bbb", HostID: "hostid2"}, StateUp, StateUp},
		{checks.CheckResult{CheckID: "cid1", CheckHostID: "bbb", HostID: "hostid1", Error: "error"}, StateUp, StateUp},
		{checks.CheckResult{CheckID: "cid1", CheckHostID: "bbb", HostID: "hostid2"}, StateUp, StateUp},
		{checks.CheckResult{CheckID: "cid1", CheckHostID: "bbb", HostID: "hostid1", Error: "error"}, StateUp, StateUp},
		{checks.CheckResult{CheckID: "cid1", CheckHostID: "bbb", HostID: "hostid2"}, StateUp, StateUp},
		{checks.CheckResult{CheckID: "cid1", CheckHostID: "bbb", HostID: "hostid1"}, StateUp, StateUp},
		{checks.CheckResult{CheckID: "cid1", CheckHostID: "bbb", HostID: "hostid2", Error: "error"}, StateUp, StateUp},
		{checks.CheckResult{CheckID: "cid1", CheckHostID: "bbb", HostID: "hostid1"}, StateUp, StateUp},
		{checks.CheckResult{CheckID: "cid1", CheckHostID: "bbb", HostID: "hostid2"}, StateUp, StateUp},
		{checks.CheckResult{CheckID: "cid1", CheckHostID: "bbb", HostID: "hostid1", Error: "error"}, StateUp, StateUp},
		{checks.CheckResult{CheckID: "cid1", CheckHostID: "bbb", HostID: "hostid2", Error: "error"}, StateDown, StateDown},
		{checks.CheckResult{CheckID: "cid1", CheckHostID: "bbb", HostID: "hostid1", Error: "error"}, StateDown, StateDown},
		{checks.CheckResult{CheckID: "cid1", CheckHostID: "bbb", HostID: "hostid2", Error: "error"}, StateDown, StateDown},
		{checks.CheckResult{CheckID: "cid1", CheckHostID: "bbb", HostID: "hostid1"}, StateDown, StateDown},
		{checks.CheckResult{CheckID: "cid1", CheckHostID: "bbb", HostID: "hostid2"}, StateDown, StateDown},
		{checks.CheckResult{CheckID: "cid1", CheckHostID: "bbb", HostID: "hostid1", Error: "error"}, StateDown, StateDown},
		{checks.CheckResult{CheckID: "cid1", CheckHostID: "bbb", HostID: "hostid2"}, StateUp, StateUp},
		{checks.CheckResult{CheckID: "cid1", CheckHostID: "bbb", HostID: "hostid1"}, StateUp, StateUp},
	}

	for i, c := range cases {
		err := db.Save(&c.in)
		if err != nil {
			t.Fatalf("Save() failed: %s", err.Error())
		}

		e1, _ := e.Evaluate(&c.in)
		if e1.State != c.state1 {
			t.Fatalf("Evaluate() [%d] concluded wrong state. Got %s, expected %s", i, e1.State.ColorString(), c.state1.ColorString())
		}

		e2, err := e.evaluteHost(e1)
		if err == nil && e2.State != c.state2 {
			t.Fatalf("evaluteHost() [%d] concluded wrong state. Got %s, expected %s", i, e2.State.ColorString(), c.state2.ColorString())
		}
	}
}
