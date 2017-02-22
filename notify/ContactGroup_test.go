package notify

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/gansoi/gansoi/boltdb"
	"github.com/gansoi/gansoi/database"
)

func TestLoadContactGroupFail(t *testing.T) {
	db := boltdb.NewTestStore()

	g, err := LoadContactGroup(db, "nonexisting")
	if err == nil {
		t.Fatalf("LoadContact() failed to catch error")
	}

	if g != nil {
		t.Fatalf("LoadCOntact() returned non-nil on error")
	}
}

func TestLoadContactGroup(t *testing.T) {
	db := boltdb.NewTestStore()

	g1 := &ContactGroup{
		Members: nil,
	}
	g1.ID = "buh"

	err := db.Save(g1)
	if err != nil {
		t.Fatalf("db.Save() failed: %s", err.Error())
	}

	g2, err := LoadContactGroup(db, "buh")
	if err != nil {
		t.Fatalf("LoadContactGroup() failed: %s", err.Error())
	}

	if g2 == nil {
		t.Fatalf("LoadContactGroup() returned nil without error")
	}

	if !reflect.DeepEqual(g1, g2) {
		t.Fatalf("Did not load the same Contact as saevd, got %v, expected %v", g2, g1)
	}
}

func TestGetContacts(t *testing.T) {
	db := boltdb.NewTestStore()

	g := &ContactGroup{}
	g.ID = "buh"

	contacts, err := g.GetContacts(db)
	if err != nil {
		t.Fatalf("GetContacts() failed: %s", err.Error())
	}

	if len(contacts) != 0 {
		t.Fatalf("len != 0")
	}

	g.Members = []string{"m1", "m2"}
	contacts, err = g.GetContacts(db)
	if err != database.ErrNotFound {
		t.Fatalf("GetContacts() failed to err on missing contacts")
	}

	if len(contacts) != 0 {
		t.Fatalf("len != 0")
	}

	c := &Contact{
		Notifier:  "none",
		Arguments: json.RawMessage("{}"),
	}

	c.ID = "m1"
	db.Save(c)
	c.ID = "m2"
	db.Save(c)

	contacts, err = g.GetContacts(db)
	if err != nil {
		t.Fatalf("GetContacts() failed: %s", err.Error())
	}

	if len(contacts) != 2 {
		t.Fatalf("len != 2")
	}
}

func TestContactGroupValidate(t *testing.T) {
	db := boltdb.NewTestStore()

	contact := &Contact{Name: "Name", Arguments: json.RawMessage("{}")}
	db.Save(contact)

	cases := []struct {
		in  *ContactGroup
		err bool
	}{
		{&ContactGroup{}, true},
		{&ContactGroup{Name: "name"}, false},
		{&ContactGroup{Name: "name", Members: []string{"hopsa"}}, true},
		{&ContactGroup{Name: "name", Members: []string{contact.ID}}, false},
		{&ContactGroup{Name: "name", Members: []string{contact.ID, "hopsa"}}, true},
	}

	for i, c := range cases {
		err := c.in.Validate(db)

		// Got no error, expected error
		if err == nil && c.err {
			t.Fatalf("%d: Failed to catch validation error in %+v", i, c.in)
		}

		// Got error, expected none
		if err != nil && !c.err {
			t.Fatalf("%d: Wrongly catched validation error in %+v (%s)", i, c.in, err.Error())
		}
	}
}
