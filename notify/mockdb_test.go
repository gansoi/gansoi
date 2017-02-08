package notify

import (
	"os"
	"testing"

	"github.com/gansoi/gansoi/boltdb"
)

func newDB(t *testing.T) *boltdb.BoltStore {
	path := "/dev/shm/mockdb"
	db, err := boltdb.NewBoltStore(path)
	if err != nil {
		t.Fatalf("NewBoltStore() failed: %s", err.Error())
	}

	err = os.Remove(path)
	if err != nil {
		t.Fatalf("os.Remove() failed: %s", err.Error())
	}

	return db
}
