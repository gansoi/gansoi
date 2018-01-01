package config

import (
	"errors"
	"testing"

	"github.com/gansoi/gansoi/boltdb"
	"github.com/gansoi/gansoi/checks"
	"github.com/gansoi/gansoi/notify"
	"github.com/gansoi/gansoi/transports/ssh"
)

func TestLoadFromFile(t *testing.T) {
	var conf Configuration
	err := conf.LoadFromFile("testdata/config.yml")
	if err != nil {
		t.Fatalf("LoadFromFile() failed: %s", err.Error())
	}

	err = conf.LoadFromFile("testdata/config.yml-NONEXISTING")
	if err == nil {
		t.Fatalf("LoadFromFile() did not fail for nonexisting file")
	}

	err = conf.LoadFromFile("testdata/malformed-config.yml")
	if err == nil {
		t.Fatalf("LoadFromFile() did not fail for malformed input")
	}
}

func TestDefaults(t *testing.T) {
	conf := NewConfiguration()
	if conf.Bind != ":4934" {
		t.Fatalf("Wrong default")
	}

	if !conf.HTTP.TLS {
		t.Fatalf("Configuration should use TLS as default.")
	}

	if conf.DataDir == "" {
		t.Fatalf("No default datadir")
	}
}

type writer struct {
	saves int
	err   error
}

func (w *writer) Save(data interface{}) error {
	w.saves++

	return w.err
}

func (w *writer) Delete(data interface{}) error {
	return w.err
}

func TestSaveChecks(t *testing.T) {
	w := &writer{}

	conf := NewConfiguration()
	conf.LoadFromFile("testdata/checks.yml")

	err := conf.SaveChecks(w)
	if err != nil {
		t.Errorf("SaveChecks() returned an error: %s", err.Error())
	}

	if w.saves != 1 {
		t.Errorf("Something went wrong, %d checks saved", w.saves)
	}

	w.err = errors.New("error")
	err = conf.SaveChecks(w)
	if err == nil {
		t.Errorf("SaveChecks() did not return an error")
	}
}

func TestSaveHosts(t *testing.T) {
	w := &writer{}

	conf := NewConfiguration()
	conf.LoadFromFile("testdata/hosts.yml")

	err := conf.SaveHosts(w)
	if err != nil {
		t.Errorf("SaveHosts() returned an error: %s", err.Error())
	}

	if w.saves != 1 {
		t.Errorf("Something went wrong, %d hosts saved", w.saves)
	}

	w.err = errors.New("error")
	err = conf.SaveHosts(w)
	if err == nil {
		t.Errorf("SaveHosts() did not return an error")
	}
}

func TestSaveContactGroups(t *testing.T) {
	w := &writer{}

	conf := NewConfiguration()
	conf.LoadFromFile("testdata/contactgroups.yml")

	err := conf.SaveContactGroups(w)
	if err != nil {
		t.Errorf("SaveContactGroups() returned an error: %s", err.Error())
	}

	if w.saves != 1 {
		t.Errorf("Something went wrong, %d contactgroups saved", w.saves)
	}

	w.err = errors.New("error")
	err = conf.SaveContactGroups(w)
	if err == nil {
		t.Errorf("SaveContactGroups() did not return an error")
	}
}

func TestSaveContacts(t *testing.T) {
	w := &writer{}

	conf := NewConfiguration()
	conf.LoadFromFile("testdata/contacts.yml")

	err := conf.SaveContacts(w)
	if err != nil {
		t.Errorf("SaveContacts() returned an error: %s", err.Error())
	}

	if w.saves != 1 {
		t.Errorf("Something went wrong, %d contacts saved", w.saves)
	}

	w.err = errors.New("error")
	err = conf.SaveContacts(w)
	if err == nil {
		t.Errorf("SaveContacts() did not return an error")
	}
}

func TestDeleteUnknownSeeds(t *testing.T) {
	db := boltdb.NewTestStore()
	defer db.Close()

	check := &checks.Check{}
	err := db.Save(check)
	if err != nil {
		t.Fatalf("Failed to save check: %s", err.Error())
	}

	host := &ssh.SSH{}
	err = db.Save(host)
	if err != nil {
		t.Fatalf("Failed to save host: %s", err.Error())
	}

	contactgroup := &notify.ContactGroup{}
	err = db.Save(contactgroup)
	if err != nil {
		t.Fatalf("Failed to save contactgroup: %s", err.Error())
	}

	contact := &notify.Contact{}
	err = db.Save(contact)
	if err != nil {
		t.Fatalf("Failed to save contact: %s", err.Error())
	}

	conf := NewConfiguration()
	conf.DeleteUnknownSeeds(db)

	var checks []*checks.Check
	db.All(&checks, -1, 0, false)
	if len(checks) != 0 {
		t.Errorf("Wrong number of checks left in database, got %d, expected 0", len(checks))
	}

	var hosts []*ssh.SSH
	db.All(&hosts, -1, 0, false)
	if len(hosts) != 0 {
		t.Errorf("Wrong number of hosts left in database, got %d, expected 0", len(hosts))
	}

	var contacts []*notify.Contact
	db.All(&contacts, -1, 0, false)
	if len(contacts) != 0 {
		t.Errorf("Wrong number of contacts left in database, got %d, expected 0", len(contacts))
	}

	var contactgroups []*notify.ContactGroup
	db.All(&contactgroups, -1, 0, false)
	if len(contactgroups) != 0 {
		t.Errorf("Wrong number of contactgroups left in database, got %d, expected 0", len(contactgroups))
	}
}
