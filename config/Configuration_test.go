package config

import (
	"testing"
)

func TestLoadFromFile(t *testing.T) {
	var conf Configuration
	err := conf.LoadFromFile("testdata/config.conf")
	if err != nil {
		t.Fatalf("LoadFromFile() failed: %s", err.Error())
	}

	err = conf.LoadFromFile("testdata/config.conf-NONEXISTING")
	if err == nil {
		t.Fatalf("LoadFromFile() did not fail for nonexisting file")
	}

	err = conf.LoadFromFile("testdata/malformed-config.conf")
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
