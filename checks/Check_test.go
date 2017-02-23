package checks

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/gansoi/gansoi/boltdb"
	"github.com/gansoi/gansoi/plugins"
)

type (
	mockAgent struct {
		ReturnError bool `json:"return_error"`
		Panic       bool `json:"panic"`
	}
)

func (m *mockAgent) Check(result plugins.AgentResult) error {
	if m.ReturnError {
		return errors.New("error")
	}

	if m.Panic {
		panic("panic")
	}

	result.AddValue("ran", true)

	return nil
}

func init() {
	plugins.RegisterAgent("mock", mockAgent{})
}

func TestCheckJsonInvalid(t *testing.T) {
	cases := []string{
		`{"id": 12}`,
		`{"agent": "nonexisting"}`,
		`{"agent": "mock", "arguments": "wrongtype"}`,
	}

	var check Check
	for _, input := range cases {
		err := json.Unmarshal([]byte(input), &check)

		if err == nil {
			t.Fatalf("Unmarshal did not catch broken json '%s'", input)
		}
	}
}

func TestCheckJson(t *testing.T) {
	input := []byte(`{
    	"id": "tester",
    	"agent": "mock",
    	"arguments": {
    	}
    }`)

	var check Check
	err := json.Unmarshal(input, &check)
	if err != nil {
		t.Fatalf("Unmarshal failed: %s", err.Error())
	}

	if check.ID != "tester" {
		t.Fatalf("ID is not 'test', (got %s)", check.ID)
	}
}

func TestRunCheck(t *testing.T) {
	input := []byte(`{
    	"id": "tester",
    	"agent": "mock",
    	"arguments": {},
        "expressions": ["ran == true"]
    }`)

	var check Check
	err := json.Unmarshal(input, &check)
	if err != nil {
		t.Fatalf("Unmarshal failed: %s", err.Error())
	}

	result := RunCheck(&check)
	if result.Results["ran"] != true {
		t.Fatalf("Check failed to run")
	}
}

func TestRunCheckError(t *testing.T) {
	cases := []string{
		`{"agent": "mock", "arguments": {"return_error": true}, "expressions": ["ran == true"]}`,
		`{"agent": "mock", "arguments": {}, "expressions": ["<<"]}`,
		`{"agent": "mock", "arguments": {}, "expressions": ["nonexisting < 10"]}`,
		`{"agent": "mock", "arguments": {}, "expressions": ["ran == false"]}`,
		`{"agent": "mock", "arguments": {"panic": true}, "expressions": []}`,
	}

	var check Check
	for _, input := range cases {
		err := json.Unmarshal([]byte(input), &check)
		if err != nil {
			t.Fatalf("Unmarshal failed: %s", err.Error())
		}

		result := RunCheck(&check)
		if result.Error == "" {
			t.Fatalf("Failed to return error for '%s'", input)
		}
	}
}

func TestCheckValidate(t *testing.T) {
	db := boltdb.NewTestStore()

	cases := []struct {
		in  *Check
		err bool
	}{
		{&Check{}, true},
		{&Check{Name: "name"}, true},
		{&Check{AgentID: "agent"}, true},
		{&Check{Name: "name", AgentID: "agent"}, false},
	}

	for i, c := range cases {
		err := c.in.Validate(db)

		// Got no error, expected error
		if err == nil && c.err {
			t.Fatalf("%d: Failed to catch validation error in %+v", i, c.in)
		}

		// Got error, expected none
		if err != nil && !c.err {
			t.Fatalf("%d: Wrongly catched validation error in %+v", i, c.in)
		}
	}
}
