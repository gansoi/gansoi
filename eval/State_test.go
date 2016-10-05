package eval

import (
	"encoding/json"
	"testing"
)

func TestValid(t *testing.T) {
	cases := []struct {
		input    State
		expected bool
	}{
		{StateUnknown, true},
		{StateUp, true},
		{StateDegraded, true},
		{StateDown, true},
		{State(39), false},
	}

	for _, dat := range cases {
		result := dat.input.Valid()

		if result != dat.expected {
			t.Fatalf("Failed to validate '%s' correct. Got %v, expected %v", dat.input, result, dat.expected)
		}
	}
}

func TestMarshalJSON(t *testing.T) {
	cases := []struct {
		input  State
		output string
	}{
		{StateUnknown, `""`},
		{StateUp, `"up"`},
		{StateDegraded, `"degraded"`},
		{StateDown, `"down"`},
		{State(39), `""`},
	}

	for _, dat := range cases {
		b, err := json.Marshal(dat.input)
		if err != nil {
			t.Fatalf("Got error on %s: %s", dat.input, err.Error())
		}

		if string(b) != dat.output {
			t.Fatalf("Got wrong output: %s, expected %s", string(b), dat.output)
		}
	}
}

func TestUnmarshalJSON(t *testing.T) {
	j := `{
        "s1": "up",
        "s2": "down",
        "s3": "",
        "s4": "degraded"
        }`

	out := make(map[string]State)

	err := json.Unmarshal([]byte(j), &out)
	if err != nil {
		t.Fatalf("JSON unmarshal failed: %s", err.Error())
	}

	if out["s1"] != StateUp || out["s2"] != StateDown || out["s3"] != StateUnknown || out["s4"] != StateDegraded {
		t.Fatalf("Failed to decode JSON properly")
	}
}
