package checks

import (
	"testing"
	"time"

	"github.com/gansoi/gansoi/boltdb"
	"github.com/gansoi/gansoi/database"
)

func TestPostApply(t *testing.T) {
	db := boltdb.NewTestStore()

	s, err := newMetaStore(db)
	if err != nil {
		t.Fatalf("newMetaStore() returned an error: %s", err.Error())
	}

	c := &Check{
		Hosts:    []string{"hostid"},
		Interval: time.Second * 61,
	}

	s.PostApply(false, database.CommandSave, c, nil)
}

func TestPostApplyNil(t *testing.T) {
	db := boltdb.NewTestStore()

	s, _ := newMetaStore(db)
	s.PostApply(false, database.CommandSave, nil, nil)
}

func TestAddCheck(t *testing.T) {
	db := boltdb.NewTestStore()

	s, _ := newMetaStore(db)

	c := &Check{
		Hosts:    []string{"hostid"},
		Interval: time.Second * 61,
	}
	s.addCheck(time.Now(), c)
}

func TestRemoteCheck(t *testing.T) {
	db := boltdb.NewTestStore()

	s, _ := newMetaStore(db)

	c := &Check{
		Hosts:    []string{"hostid"},
		Interval: time.Second * 61,
	}
	s.addCheck(time.Now(), c)
	s.removeCheck(c)
}

var _ database.Listener = (*MetaStore)(nil)
