package database

import (
	"testing"
)

func TestCommandString(t *testing.T) {
	cases := []struct {
		in       Command
		expected string
	}{
		{CommandSave, "save"},
		{CommandDelete, "delete"},
		{Command(200), "n/a"},
	}

	for _, dat := range cases {
		result := dat.in.String()
		if result != dat.expected {
			t.Fatalf("String() failed, got %s, expected %s", result, dat.expected)
		}
	}
}
