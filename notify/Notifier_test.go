package notify

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

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

	mockNotifier struct {
	}
)

var (
	notifyMessage string
)

func init() {
	plugins.RegisterAgent("mock", mock{})
	plugins.RegisterNotifier("mockn", mockNotifier{})
}

func (m *mock) Check(result plugins.AgentResult) error {
	if m.Err {
		return errors.New("error")
	}

	return nil
}

func (m mockNotifier) Notify(text string) error {
	notifyMessage = text

	return nil
}

func TestGotEvaluation(t *testing.T) {
	db := boltdb.NewTestStore()

	contact := &Contact{
		Name:     "testcontact",
		Notifier: "mockn",
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

	// This should not fail :)
	n.PostApply(false, database.CommandSave, nil)

	timeline := []struct {
		err             bool
		expectedMessage string
	}{
		{false, ""},
		{false, ""},
		{false, ""},
		{false, ""},
		{false, ""},
		{false, ""},
		{false, ""},
		{false, ""},
		{false, ""},
		{false, ""},
		{false, ""},
		{true, ""},
		{true, ""},
		{true, "Down"},
		{true, ""},
		{true, ""},
		{true, ""},
		{true, ""},
		{true, ""},
		{true, ""},
		{true, ""},
		{true, ""},
		{true, ""},
		{true, ""},
		{true, ""},
		{true, ""},
		{true, ""},
		{false, ""},
		{false, ""},
		{false, "Up"},
		{false, ""},
		{false, ""},
		{true, ""},
		{false, ""},
		{false, ""},
		{true, ""},
		{false, ""},
		{false, ""},
	}

	for i, c := range timeline {
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

		time.Sleep(time.Millisecond * 15)
		e.PostApply(true, database.CommandSave, result)
		evaluation, err := eval.LatestEvaluation(db, result)
		if err != nil {
			t.Fatalf("LatestEvaluation() failed: %s", err.Error())
		}

		time.Sleep(time.Millisecond * 15)
		n.PostApply(true, database.CommandSave, evaluation)
		//		err = n.gotEvaluation(evaluation)
		if err != nil {
			t.Fatalf("gotEvaluation() failed: %s", err.Error())
		}

		if c.expectedMessage != "" && !strings.Contains(notifyMessage, c.expectedMessage) {
			t.Errorf("%d: Notification '%s' did not contain '%s' as expected", i, notifyMessage, c.expectedMessage)
		}

		if c.expectedMessage == "" && notifyMessage != "" {
			t.Errorf("%d: Got unexpected notoification: %s", i, notifyMessage)
		}

		notifyMessage = ""
	}

	result := checks.RunCheck(nil, check)
	result.CheckHostID = checks.CheckHostID(check.GetID(), "")
	result.CheckID = check.GetID()
	e.PostApply(true, database.CommandSave, result)
	evaluation, _ := eval.LatestEvaluation(db, result)
	evaluation.CheckID = "nonexisting"
	n.PostApply(true, database.CommandSave, evaluation)
}

var _ database.Listener = (*Notifier)(nil)
