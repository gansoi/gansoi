package notify

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/gansoi/gansoi/boltdb"
	"github.com/gansoi/gansoi/checks"
	"github.com/gansoi/gansoi/database"
	"github.com/gansoi/gansoi/eval"
	"github.com/gansoi/gansoi/plugins"
)

type (
	mock struct {
		Err bool `json:"err"`
	}
)

func init() {
	plugins.RegisterAgent("mock", mock{})
}

func (m *mock) Check(result plugins.AgentResult) error {
	if m.Err {
		return errors.New("error")
	}

	return nil
}

func TestGotEvaluation(t *testing.T) {
	db := boltdb.NewTestStore()

	contact := &Contact{
		Name:     "testcontact",
		Notifier: "testnotifier",
	}
	err := db.Save(contact)
	if err != nil {
		t.Fatalf("Save() failed: %s", err.Error())
	}

	group := &ContactGroup{
		Name:    "testgroup",
		Members: []string{contact.GetID()},
	}
	err = db.Save(group)
	if err != nil {
		t.Fatalf("Save() failed: %s", err.Error())
	}

	check := &checks.Check{
		Name:          "test",
		AgentID:       "mock",
		ContactGroups: []string{group.GetID(), "nonexisting"},
	}
	err = db.Save(check)
	if err != nil {
		t.Fatalf("Save() failed: %s", err.Error())
	}

	e := eval.NewEvaluator(db)
	n, _ := NewNotifier(db)

	timeline := []struct {
		err bool
	}{
		{false},
		{false},
		{false},
		{false},
		{false},
		{false},
		{false},
		{false},
		{false},
		{false},
		{false},
		{true},
		{true},
		{true},
		{true},
		{true},
		{true},
		{true},
		{true},
		{true},
		{true},
		{true},
		{true},
		{true},
		{true},
		{true},
		{true},
		{false},
		{false},
		{false},
		{false},
		{false},
		{true},
		{false},
		{false},
		{true},
		{false},
		{false},
	}

	for _, c := range timeline {
		if c.err {
			check.Arguments = json.RawMessage(`{"err": true}`)
		} else {
			check.Arguments = json.RawMessage(`{}`)
		}

		result := checks.RunCheck(nil, check)
		result.CheckHostID = checks.CheckHostID(check.GetID(), "")
		result.CheckID = check.GetID()

		err = db.Save(result)
		if err != nil {
			t.Fatalf("Save() failed: %s", err.Error())
		}

		e.PostApply(true, database.CommandSave, result)
		evaluation, err := eval.LatestEvaluation(db, result)
		if err != nil {
			t.Fatalf("LatestEvaluation() failed: %s", err.Error())
		}

		err = n.gotEvaluation(evaluation)
		if err != nil {
			t.Fatalf("gotEvaluation() failed: %s", err.Error())
		}
	}
}

var _ database.Listener = (*Notifier)(nil)
