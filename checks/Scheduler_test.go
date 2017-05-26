package checks

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/gansoi/gansoi/boltdb"
	"github.com/gansoi/gansoi/database"
	"github.com/gansoi/gansoi/stats"
)

type (
	failDB struct {
		err error
	}
)

func (f *failDB) Save(_ interface{}) error {
	return f.err
}

func (f *failDB) One(_ string, _ interface{}, _ interface{}) error {
	return f.err
}

func (f *failDB) All(_ interface{}, _ int, _ int, _ bool) error {
	return f.err
}

func (f *failDB) Find(field string, value interface{}, to interface{}, limit int, skip int, reverse bool) error {
	return f.err
}

func (f *failDB) Delete(data interface{}) error {
	return f.err
}

func (f *failDB) RegisterListener(listener database.Listener) {
}

func TestNewScheduler(t *testing.T) {
	db := boltdb.NewTestStore()

	s := NewScheduler(db, "test")
	s.Run()
	time.Sleep(time.Millisecond * 500)
	s.Stop()
}

func TestSchedulerLoop(t *testing.T) {
	c := &Check{
		AgentID:   "mock",
		Interval:  time.Millisecond * 100,
		Arguments: json.RawMessage("{}"),
	}
	c.ID = "hello"

	db := boltdb.NewTestStore()
	db.Save(c)

	s := NewScheduler(db, "test")
	s.Run()
	time.Sleep(time.Millisecond * 500)
	s.Stop()
}

func TestSchedulerLoopCheck(t *testing.T) {
	c := &Check{
		Interval:  time.Millisecond * 1,
		AgentID:   "mock",
		Arguments: json.RawMessage(`{"panic": false}`),
	}
	c.ID = "hello"

	db := boltdb.NewTestStore()
	err := db.Save(c)
	if err != nil {
		t.Fatalf("Save() failed: %s", err.Error())
	}

	s := NewScheduler(db, "test")
	s.Run()
	time.Sleep(time.Millisecond * 500)
	s.Stop()

	var results []*CheckResult
	err = db.All(&results, -1, 0, false)
	if err != nil {
		t.Fatalf("All() failed: %s", err.Error())
	}

	if len(results) == 0 {
		t.Fatalf("Did not run test at all")
	}
}

func TestSchedulerLoopCheckOverrun(t *testing.T) {
	stats.CounterSet("scheduler_inflight_overrun", 0)

	c := &Check{
		Interval:  time.Millisecond * 1,
		AgentID:   "mock",
		Arguments: json.RawMessage(`{"delay":510000000}`),
	}
	c.ID = "hello"

	db := boltdb.NewTestStore()
	err := db.Save(c)
	if err != nil {
		t.Fatalf("Save() failed: %s", err.Error())
	}

	s := NewScheduler(db, "test")
	s.Run()
	time.Sleep(time.Millisecond * 1500)
	s.Stop()

	var results []*CheckResult
	err = db.All(&results, -1, 0, false)
	if err != nil {
		t.Fatalf("All() failed: %s", err.Error())
	}

	if len(results) == 0 {
		t.Fatalf("Did not run test at all")
	}

	overruns := stats.CounterGet("scheduler_inflight_overrun")
	if overruns == 0 {
		t.Fatalf("Scheduler did not detect overrun")
	}
}

func TestSchedulerRunCheck(t *testing.T) {
	c := Check{
		Interval:  time.Millisecond * 100,
		AgentID:   "mock",
		Arguments: []byte("{}"),
	}

	db := boltdb.NewTestStore()
	s := NewScheduler(db, "test")

	clock := time.Now()

	meta := &checkMeta{
		check: c,
		key:   &metaKey{},
	}
	result := s.runCheck(clock, meta)
	ran, _ := result.Results["ran"].(bool)

	if result.Error != "" {
		t.Fatalf("runCheck() returned an error: %s", result.Error)
	}

	if !ran {
		t.Fatalf("runCheck() did not execute the check")
	}

	// Now try to panic.
	c.Arguments = []byte(`{"panic": true}`)
	s.runCheck(clock, meta)
}

func TestSpinFail(t *testing.T) {
	db := &failDB{
		err: errors.New("failed on purpose"),
	}

	s := NewScheduler(db, "test")
	s.spin(time.Now())
}
