package notify

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/gansoi/gansoi/boltdb"
)

func TestContactNotify(t *testing.T) {
	c := Contact{
		Notifier:  "mockagent",
		Arguments: json.RawMessage(`{"PleaseReturn": ""}`),
	}
	c.ID = "test"

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
	db := boltdb.NewTestStore()

	c, err := LoadContact(db, "nonexisting")
	if err == nil {
		t.Fatalf("LoadContact() failed to catch error")
	}

	if c != nil {
		t.Fatalf("LoadCOntact() returned non-nil on error")
	}
}

func TestLoadContact(t *testing.T) {
	db := boltdb.NewTestStore()

	c1 := &Contact{
		Notifier:  "none",
		Arguments: json.RawMessage("{}"),
	}
	c1.ID = "buh"

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

func TestContactValidate(t *testing.T) {
	db := boltdb.NewTestStore()

	cases := []struct {
		in  *Contact
		err bool
	}{
		{&Contact{}, true},
		{&Contact{Name: "name"}, true},
		{&Contact{Notifier: "notifier"}, true},
		{&Contact{Name: "name", Notifier: "notifier"}, false},
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
