package boltdb

import (
	"os"
	"testing"
	"time"

	"github.com/gansoi/gansoi/database"
	"github.com/hashicorp/raft"
)

var (
	db *BoltStore
)

const (
	path = "/dev/shm/gansoi-test.db"
)

type (
	TestDb struct {
		*BoltStore
	}

	data struct {
		ID string `storm:"id"`
		A  string
	}
)

func init() {
	database.RegisterType(data{})
}

func newTestDb() *TestDb {
	var err error
	db, err = NewBoltStore(path)
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

func TestBoltStoreSave(t *testing.T) {
	db := newTestDb()
	defer db.clean()

	d := data{
		ID: "bah",
		A:  "hello",
	}

	err := db.Save(&d)
	if err != nil {
		t.Fatalf("Save() failed: %s", err.Error())
	}
}

func TestBoltStoreOne(t *testing.T) {
	db := newTestDb()
	defer db.clean()

	d := data{
		ID: "bah",
		A:  "hello",
	}

	err := db.Save(&d)
	if err != nil {
		t.Fatalf("Save() failed: %s", err.Error())
	}

	err = db.One("ID", "bah", &d)
	if err != nil {
		t.Fatalf("One() failed: %s", err.Error())
	}

	err = db.One("ID", "bah2", &d)
	if err != database.ErrNotFound {
		t.Fatalf("One() did not return correct error for record not found, returned: %s", err.Error())
	}
}

func TestBoltStoreAll(t *testing.T) {
	db := newTestDb()
	defer db.clean()

	d := data{
		ID: "bah",
		A:  "hello",
	}

	var all []data
	err := db.All(&all, -1, 0, false)
	if err != nil {
		t.Fatalf("All() failed: %s", err.Error())
	}

	if len(all) != 0 {
		t.Fatalf("All() returned wrong number of results, expected 0, got %d", len(all))
	}

	err = db.Save(&d)
	if err != nil {
		t.Fatalf("Save() failed: %s", err.Error())
	}

	err = db.All(&all, -1, 0, false)
	if err != nil {
		t.Fatalf("All() failed: %s", err.Error())
	}

	if len(all) != 1 {
		t.Fatalf("All() returned wrong number of results, expected 1, got %d", len(all))
	}

	if all[0].ID != d.ID || all[0].A != d.A {
		t.Fatalf("All() returned wrong result %v != %v", all[0], d)
	}
}

func TestBoltStoreStorm(t *testing.T) {
	db := newTestDb()
	defer db.clean()

	if db.Storm().Bolt.Path() != path {
		t.Fatalf("Something is wrong with the underlying Storm/Bolt storage")
	}
}

func TestProcessLogEntry(t *testing.T) {
	db := newTestDb()
	defer db.clean()

	d := data{
		ID: "bah",
		A:  "hello",
	}

	save := database.NewLogEntry(database.CommandSave, d)
	err := db.ProcessLogEntry(save)
	if err != nil {
		t.Fatalf("ProcessLogEntry() failed: %s", err.Error())
	}

	err = db.One("ID", "bah", &d)
	if err != nil {
		t.Fatalf("One() failed: %s", err.Error())
	}

	del := database.NewLogEntry(database.CommandDelete, d)
	err = db.ProcessLogEntry(del)
	if err != nil {
		t.Fatalf("ProcessLogEntry() failed: %s", err.Error())
	}

	err = db.One("ID", "bah", &d)
	if err == nil {
		t.Fatalf("CommandDelete failed to delete")
	}

	broken := database.NewLogEntry(76, d)
	err = db.ProcessLogEntry(broken)
	if err == nil {
		t.Fatalf("ProcessLogEntry() failed to detect unknown command")
	}

	// Give the store a chance to run goroutines.
	time.Sleep(time.Millisecond * 100)
}

func TestBoltStoreApply(t *testing.T) {
	db := newTestDb()
	defer db.clean()

	d := data{
		ID: "bah",
		A:  "hello",
	}

	entry := database.NewLogEntry(database.CommandSave, &d)
	l := &raft.Log{}
	l.Type = raft.LogCommand

	ret := db.Apply(l)
	if ret != nil {
		t.Fatalf("Apply() failed to recognize broken log entry")
	}

	l.Data = entry.Byte()
	ret = db.Apply(l)

	err, ok := ret.(error)
	if ok && err != nil {
		t.Fatalf("Apply() failed: %s", err.Error())
	}
}

func TestDatabaseWriteTo(t *testing.T) {
	backupPath := path + ".backup"

	db := newTestDb()

	d := data{
		ID: "bah",
		A:  "hello",
	}

	err := db.Save(&d)
	if err != nil {
		t.Fatalf("Save() failed: %s", err.Error())
	}

	f, err := os.OpenFile(backupPath, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		t.Fatalf("os.Open() failed: %s", err.Error())
	}

	n, err := db.WriteTo(f)
	if err != nil {
		t.Fatalf("WriteTo() failed: %s", err.Error())
	}

	if n < 1 {
		t.Fatal("WriteTo() saved too few bytes")
	}

	db.clean()

	os.Rename(backupPath, path)

	db = newTestDb()
	defer db.clean()

	var dd data

	err = db.One("ID", "bah", &dd)
	if err != nil {
		t.Fatalf("One() failed: %s", err.Error())
	}

	if d.ID != dd.ID || d.A != dd.A {
		t.Fatalf("Backup returned wrong data: %v Should be: %v", dd, d)
	}
}

// Make sure we implement the needed interface.
var _ database.Database = (*BoltStore)(nil)
