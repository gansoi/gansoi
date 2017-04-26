package database

import (
	"testing"
)

type (
	typed struct {
		Object
	}
)

func TestSetID(t *testing.T) {
	var buh typed

	buh.SetID()

	if buh.ID == "" {
		t.Fatalf("SetID failed to set ID")
	}

	var buh2 Object

	buh2.SetID()
	if buh2.ID == "" {
		t.Fatalf("SetID failed to set ID")
	}
}

func TestGetID(t *testing.T) {
	var buh typed

	if buh.GetID() != "" {
		t.Errorf("GetID() returned non-empty ID")
	}

	buh.Object.ID = "hellohello"

	if buh.GetID() != "hellohello" {
		t.Errorf("GetID() returned wrong ID")
	}
}

// Make sure we implement the needed interface.
var _ IDSetter = (*Object)(nil)
