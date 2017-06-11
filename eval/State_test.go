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

func TestColorString(t *testing.T) {
	cases := []struct {
		input  State
		output string
	}{
		{StateUnknown, blue + "Unknown" + reset},
		{StateUp, green + "Up" + reset},
		{StateDegraded, yellow + "Degraded" + reset},
		{StateDown, red + "Down" + reset},
		{State(39), "" + reset},
	}

	for _, dat := range cases {
		out := dat.input.ColorString()

		if out != dat.output {
			t.Fatalf("Got wrong output: %s, expected %s", out, dat.output)
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

	err = json.Unmarshal([]byte(`"hello"`), nil)
	if err == nil {
		t.Fatalf("Failed to catch invalid input")
	}
}

func TestUnmarshalJSONInvalid(t *testing.T) {
	cases := []string{
		`"invalid"`,
		`"up`,
		`2`,
		`"`,
		``,
	}

	var s State

	for _, j := range cases {
		err := json.Unmarshal([]byte(j), &s)
		if s != StateUnknown {
			t.Fatalf("Output changed by invalid input: %s", j)
		}

		if err == nil {
			t.Fatalf("Failed to catch invalid input: %s", j)
		}

	}
}
