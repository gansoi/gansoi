package notify

import (
	"encoding/json"
	"testing"
)

func TestContactNotify(t *testing.T) {
	c := Contact{
		ID:        "test",
		Notifier:  "mockagent",
		Arguments: json.RawMessage(`{"PleaseReturn": ""}`),
	}

	err := c.Notify("some message again")
	if err != nil {
		t.Fatalf("Notify() returned error: %s", err.Error())
	}

	c.Arguments = json.RawMessage(`{"PleaseReturn": "hello"}`)
	err = c.Notify("some message again")
	if err == nil {
		t.Fatalf("Notify() didn't return expected error")
	}

	if err.Error() != "hello" {
		t.Fatalf("Notify() returned unexpected error: %s", err.Error())
	}

	c.Arguments = json.RawMessage(`{invalid json`)
	err = c.Notify("invalid json test")
	if err == nil {
		t.Fatalf("Notify() didn't return expected error for malformed Arguments")
	}
}
