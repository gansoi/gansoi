package eval

import (
	"testing"
	"time"

	"github.com/gansoi/gansoi/boltdb"
	"github.com/gansoi/gansoi/checks"
	"github.com/gansoi/gansoi/database"
)

func TestLatestEvaluation(t *testing.T) {
	db := boltdb.NewTestStore()
	result := &checks.CheckResult{
		CheckHostID: "hello::",
	}

	_, err := LatestEvaluation(db, result)
	if err == nil {
		t.Fatalf("LatestEvaluation() did not fail when for zero results")
	}

	eval1 := &Evaluation{
		CheckHostID: "hello::",
	}

	err = db.Save(eval1)
	if err != nil {
		t.Fatalf("Save() failed: %s", err.Error())
	}

	_, err = LatestEvaluation(db, result)
	if err != nil && err != database.ErrNotFound {
		t.Fatalf("LatestEvaluation() failed: %s", err.Error())
	}

	_, err = LatestEvaluation(db, &checks.CheckResult{})
	if err == nil {
		t.Fatalf("LatestEvaluation() did not fail when for empty query")
	}
}

func TestFinal(t *testing.T) {
	now := time.Now()
	result := &checks.CheckResult{
		CheckID:     "hello",
		CheckHostID: "hello::",
	}

	eval := NewEvaluation(now, result)
	if eval.Final() == false {
		t.Errorf("Final() returned wrong result")
	}
}
