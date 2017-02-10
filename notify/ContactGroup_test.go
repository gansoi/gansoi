package notify

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/gansoi/gansoi/database"
)

func TestLoadContactGroupFail(t *testing.T) {
	db := newDB(t)
	defer db.Close()

	g, err := LoadContactGroup(db, "nonexisting")
	if err == nil {
		t.Fatalf("LoadContact() failed to catch error")
	}

	if g != nil {
		t.Fatalf("LoadCOntact() returned non-nil on error")
	}
}

func TestLoadContactGroup(t *testing.T) {
	db := newDB(t)
	defer db.Close()

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
	db := newDB(t)
	defer db.Close()

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
