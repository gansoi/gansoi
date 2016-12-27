package boltdb

import (
	"os"
	"testing"

	"github.com/gansoi/gansoi/database"
)

var (
	db *BoltStore
)

type (
	TestDb struct {
		*BoltStore
	}
)

func newTestDb() *TestDb {
	var err error
	db, err = NewBoltStore("/dev/shm/gansoi-test.db")
	if err != nil {
		panic(err.Error())
	}

	return &TestDb{
		BoltStore: db,
	}
}

func (d *TestDb) clean() {
	err := d.Close()
	if err != nil {
		panic(err.Error())
	}
	err = os.Remove("/dev/shm/gansoi-test.db")
	if err != nil {
		panic(err.Error())
	}
}

func TestDatabaseOpen(t *testing.T) {
	db := newTestDb()
	if db == nil {
		t.Fatalf("NewDatabase() failed to open database")
	}

	db.clean()
}

func TestDatabaseOpenFail(t *testing.T) {
	db, err := NewBoltStore("/iudfhgiudfgh/iuoshdgiusfdhgiufhdg/notexisting")
	if err == nil {
		t.Fatalf("NewDatabase() failed to return an error for unexisting path")
	}

	if db != nil {
		t.Fatalf("NewDatabase() failed to return nil for unexisting path")
	}
}

var _ database.Database = (*BoltStore)(nil)
