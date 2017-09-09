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
