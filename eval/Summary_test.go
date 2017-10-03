package eval

import (
	"testing"
	"time"

	"github.com/gansoi/gansoi/boltdb"
	"github.com/gansoi/gansoi/checks"
	"github.com/gansoi/gansoi/database"
)

func TestAddEvaluation(t *testing.T) {
	db := boltdb.NewTestStore()
	s := NewSummary(db)

	result := &checks.CheckResult{CheckID: "hello", CheckHostID: "hello::"}
	e := NewEvaluation(time.Now(), result)
	s.AddEvaluation(e)

	result = &checks.CheckResult{CheckID: "hello", CheckHostID: "hello::host"}
	e = NewEvaluation(time.Now(), result)
	s.AddEvaluation(e)
}

func TestPostApply(t *testing.T) {
	db := boltdb.NewTestStore()
	s := NewSummary(db)

	s.PostApply(false, database.CommandSave, &checks.Check{Object: database.Object{ID: "id1"}})
	if len(s.checks) != 0 {
		t.Errorf("Failed to ignore non-leader post apply")
	}

	s.PostApply(true, database.CommandSave, &checks.Check{Object: database.Object{ID: "id1"}})
	if len(s.checks) != 1 {
		t.Errorf("Failed to add check")
	}

	s.PostApply(true, database.CommandDelete, &checks.Check{Object: database.Object{ID: "id1"}})
	if len(s.checks) != 0 {
		t.Errorf("Failed to remove check")
	}

	result := &checks.CheckResult{CheckID: "hello", CheckHostID: "hello::"}
	e := NewEvaluation(time.Now(), result)
	s.PostApply(true, database.CommandSave, e)

	// This should not panic.
	s.PostApply(true, database.CommandSave, nil)
}
