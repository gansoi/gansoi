package notify

import (
	"encoding/json"
	"reflect"
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

func TestLoadContactFail(t *testing.T) {
	db := newDB(t)
	defer db.Close()

	c, err := LoadContact(db, "nonexisting")
	if err == nil {
		t.Fatalf("LoadContact() failed to catch error")
	}

	if c != nil {
		t.Fatalf("LoadCOntact() returned non-nil on error")
	}
}

func TestLoadContact(t *testing.T) {
	db := newDB(t)
	defer db.Close()

	c1 := &Contact{
		ID:        "buh",
		Notifier:  "none",
		Arguments: json.RawMessage("{}"),
	}

	db.Save(c1)
	c2, err := LoadContact(db, "buh")
	if err != nil {
		t.Fatalf("LoadContact() failed: %s", err.Error())
	}

	if c2 == nil {
		t.Fatalf("LoadContact() returned nil without error")
	}

	if !reflect.DeepEqual(c1, c2) {
		t.Fatalf("Did not load the same Contact as saevd, got %v, expected %v", c2, c1)
	}
}
