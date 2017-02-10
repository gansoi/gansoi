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
