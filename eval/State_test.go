package eval

import (
	"encoding/json"
	"testing"

	"github.com/gansoi/gansoi/logger"
)

func TestValid(t *testing.T) {
	cases := []struct {
		input    State
		expected bool
	}{
		{StateUnknown, true},
		{StateUp, true},
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
		{StateUnknown, logger.Blue + "Unknown" + logger.Reset},
		{StateUp, logger.Green + "Up" + logger.Reset},
		{StateDown, logger.Red + "Down" + logger.Reset},
		{State(39), "" + logger.Reset},
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
        "s3": ""
        }`

	out := make(map[string]State)

	err := json.Unmarshal([]byte(j), &out)
	if err != nil {
		t.Fatalf("JSON unmarshal failed: %s", err.Error())
	}

	if out["s1"] != StateUp || out["s2"] != StateDown || out["s3"] != StateUnknown {
		t.Fatalf("Failed to decode JSON properly")
	}

	err = json.Unmarshal([]byte(`"hello"`), &out)
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

func TestMarshalText(t *testing.T) {
	cases := []struct {
		input  State
		output string
	}{
		{StateUnknown, "unknown"},
		{StateUp, "up"},
		{StateDown, "down"},
		{State(39), "unknown"},
	}

	for _, dat := range cases {
		b, err := dat.input.MarshalText()
		if err != nil {
			t.Fatalf("Got error on %s: %s", dat.input, err.Error())
		}

		if string(b) != dat.output {
			t.Fatalf("Got wrong output: %s, expected %s", string(b), dat.output)
		}
	}
}

func TestUnmarshalText(t *testing.T) {
	cases := []struct {
		in       []byte
		expected State
	}{
		{[]byte("up"), StateUp},
		{[]byte("unknown"), StateUnknown},
		{[]byte("down"), StateDown},
	}

	var s State
	for _, c := range cases {
		err := s.UnmarshalText(c.in)
		if err != nil {
			t.Fatalf("UnmarshalText failed: %s", err.Error())
		}
	}

	err := s.UnmarshalText([]byte(`"hello"`))
	if err == nil {
		t.Fatalf("Failed to catch invalid input")
	}
}
