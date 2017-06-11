package eval

import (
	"bytes"
	"encoding/json"
	"reflect"
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

func equal(a, b *States) bool {
	if len(*a) != len(*b) {
		return false
	}

	for i, aa := range *a {
		if aa != (*b)[i] {
			return false
		}
	}

	return true
}

func TestLast(t *testing.T) {
	cases := []struct {
		input    States
		n        int
		expected States
	}{
		{States{}, 0, States{}},
		{States{StateDown}, 0, States{}},
		{States{StateDown, StateDegraded}, 0, States{}},
		{States{StateDown, StateUp}, 1, States{StateDown}},
		{States{StateDown, StateUp}, 2, States{StateDown, StateUp}},
		{States{StateDown, StateUp}, 3, States{StateDown, StateUp}},
		{States{StateDown, StateUp}, -5, States{}},
		{States{}, -1, States{}},
	}

	for _, c := range cases {
		result := c.input.Last(c.n)
		if !equal(result, &c.expected) {
			t.Fatalf("%v.Last(%d) failed. Returned %s, expected %s", c.input, c.n, result.ColorString(), c.expected.ColorString())
		}
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
		{States{StateUnknown, StateDegraded, StateUp}, StateUnknown},
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

func TestStatesJSON(t *testing.T) {
	cases := []struct {
		input    States
		expected []byte
	}{
		{States{}, []byte("[]")},
		{States{StateDown}, []byte(`["down"]`)},
		{States{StateUp}, []byte(`["up"]`)},
		{States{StateUnknown}, []byte(`[""]`)},
		{States{StateDegraded}, []byte(`["degraded"]`)},
		{States{StateUp, StateDegraded}, []byte(`["up","degraded"]`)},
		{States{StateUp, StateUp}, []byte(`["up","up"]`)},
		{States{StateUp, StateUp, StateUp, StateDown}, []byte(`["up","up","up","down"]`)},
		{States{StateUnknown, StateDegraded, StateUp}, []byte(`["","degraded","up"]`)},
	}

	for _, dat := range cases {
		result, err := json.Marshal(dat.input)

		if err != nil {
			t.Fatalf("json.Marshall() failed: %s", err.Error())
		}

		if !bytes.Equal(result, dat.expected) {
			t.Fatalf("Failed to JSON encode [%s] correct. Got [%s], expected [%s]", dat.input.ColorString(), result, dat.expected)
		}
	}

	for _, dat := range cases {
		var result States
		err := json.Unmarshal(dat.expected, &result)

		if err != nil {
			t.Fatalf("json.Unmarshal() failed: %s", err.Error())
		}

		if !reflect.DeepEqual(result, dat.input) {
			t.Fatalf("Failed to unmarshal [%s] correct. Got [%s], expected [%s]", dat.expected, result.ColorString(), dat.input.ColorString())
		}
	}
}

func TestStatesColorString(t *testing.T) {
	cases := []struct {
		input    States
		expected int
	}{
		{States{}, 7},
		{States{StateDown}, 21},
		{States{StateUp}, 19},
		{States{StateUnknown}, 24},
		{States{StateDegraded}, 25},
		{States{StateUp, StateDegraded}, 37},
		{States{StateUp, StateUp}, 31},
		{States{StateUp, StateUp, StateUp, StateDown}, 57},
		{States{StateUnknown, StateDegraded, StateUp}, 54},
	}

	for _, dat := range cases {
		result := dat.input.ColorString()

		if len(result) != dat.expected {
			t.Fatalf("Failed to ColorString encode [%v] correct. Got [%d], expected [%d]", dat.input, len(result), dat.expected)
		}
	}
}
