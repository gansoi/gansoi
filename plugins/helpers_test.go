package plugins

import (
	"testing"
)

func mustPanic(t *testing.T) {
	if r := recover(); r == nil {
		t.Errorf("Unexisting counter did not cause a panic")
	}
}
