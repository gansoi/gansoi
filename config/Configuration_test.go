package config

import (
	"errors"
	"testing"
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

func TestSetDefaults(t *testing.T) {
	var conf Configuration
	conf.SetDefaults()
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

	var conf Configuration
	conf.SetDefaults()
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

	var conf Configuration
	conf.SetDefaults()
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

	var conf Configuration
	conf.SetDefaults()
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

	var conf Configuration
	conf.SetDefaults()
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
