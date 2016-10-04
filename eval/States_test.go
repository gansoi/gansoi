package eval

import (
	"testing"
)

func TestAdd(t *testing.T) {
	var s States

	if len(s) != 0 {
		t.Fatalf("Length of default States is not 0")
	}

	s.Add(StateDown, -1)
	if len(s) != 1 {
		t.Fatalf("Length of default States after one Add is not 1")
	}

	s.Add(StateUp, -1)
	if len(s) != 2 {
		t.Fatalf("Length of default States after two Add is not 2")
	}

	s.Add(StateUnknown, 2)
	if len(s) != 2 {
		t.Fatalf("Length of default States after clamp is not 2")
	}

	if s[0] != StateUnknown || s[1] != StateUp {
		t.Fatalf("Wrong contents after clamp, got %s, expected %s", s, States{StateUnknown, StateUp})
	}

	s.Clamp(1)
	if len(s) != 1 {
		t.Fatalf("Wrong size after clamp")
	}

	s.Clamp(-1)
	if len(s) != 1 {
		t.Fatalf("Wrong size after clamp")
	}

	s.Clamp(0)
	if len(s) != 0 {
		t.Fatalf("Wrong size after clamp")
	}
}

func TestEvaluate(t *testing.T) {
	cases := []struct {
		input    States
		expected State
	}{
		{States{}, StateUnknown},
		{States{StateDown}, StateDown},
		{States{StateUp}, StateUp},
		{States{StateUnknown}, StateUnknown},
		{States{StateDegraded}, StateDegraded},
		{States{StateUp, StateDegraded}, StateDegraded},
		{States{StateUp, StateUp}, StateUp},
		{States{StateUp, StateUp, StateUp, StateDown}, StateDegraded},
		{States{StateUnknown, StateDegraded, StateUp}, StateDegraded},
		{States{State(34)}, StateUnknown},
		{States{StateUp, StateUp, StateUp, State(35)}, StateUnknown},
	}

	for _, dat := range cases {
		result := dat.input.Reduce()

		if result != dat.expected {
			t.Fatalf("Failed to evalute '%s' correct. Got %s, expected %s", dat.input, result, dat.expected)
		}
	}
}
