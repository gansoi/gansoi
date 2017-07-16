package boltdb

import (
	"testing"

	"github.com/gansoi/gansoi/database"
	"github.com/hashicorp/raft"
)

func TestTestStoreSave(t *testing.T) {
	db := NewTestStore()

	d := data{
		A: "hello",
	}
	d.ID = "bah"

	err := db.Save(&d)
	if err != nil {
		t.Fatalf("Save() failed: %s", err.Error())
	}

	db.Delete(&d)
	db.Close()
}

func TestTestStoreSaveFail(t *testing.T) {
	db := NewTestStore()
	db.FailSave = true

	err := db.Save(nil)
	if err == nil {
		t.Fatalf("Save() failed to return error")
	}

	db.Close()
}

func TestTestStoreRegisterListener(t *testing.T) {
	db := NewTestStore()
	db.RegisterListener(nil)
	db.Close()
}

// Make sure we implement the needed interfaces.
var _ database.ReadWriter = (*TestStore)(nil)
var _ database.Broadcaster = (*TestStore)(nil)
var _ raft.FSM = (*BoltStore)(nil)
