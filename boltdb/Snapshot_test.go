package boltdb

import (
	"testing"

	"github.com/hashicorp/raft"
)

var _ raft.FSMSnapshot = (*Snapshot)(nil)

func TestSnapshotPersists(t *testing.T) {
	db := NewTestStore()

	d := data{
		A: "hello",
	}
	d.ID = "bah"

	err := db.Save(&d)
	if err != nil {
		t.Fatalf("Save() failed: %s", err.Error())
	}

	snapshot, err := db.Snapshot()
	if err != nil {
		t.Errorf("db.Snapshot() %s", err.Error())
	}

	err = snapshot.Persist(&raft.DiscardSnapshotSink{})
	if err != nil {
		t.Errorf("snapshot.Persist() %s", err.Error())
	}

	db.Close()
}

func TestSnapshotRelease(t *testing.T) {
	(&Snapshot{}).Release()
}
