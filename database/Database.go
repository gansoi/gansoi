package database

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/asdine/storm"
	"github.com/asdine/storm/index"
	"github.com/boltdb/bolt"
	"github.com/hashicorp/raft"
)

type (
	// Database is the lowest level of the gansoi database.
	Database struct {
		dbMutex sync.RWMutex
		db      *storm.DB
	}

	// Command is used to denote which operation should be carried out as a
	// result of a Raft commit.
	Command int

	// LogEntry is an entry in the Raft log (?).
	LogEntry struct {
		Command Command
		Type    string
		Value   json.RawMessage
	}
)

var (
	types     = make(map[string]reflect.Type)
	typesLock sync.Mutex
)

const (
	// CommandSave will save an object in the local database.
	CommandSave = iota

	// CommandDelete will delete an object in the local database.
	CommandDelete
)

// RegisterType will register the type with the log marshaller.
func RegisterType(v interface{}) {
	typesLock.Lock()
	defer typesLock.Unlock()

	typ := reflect.TypeOf(v)
	name := typ.String()

	_, found := types[name]
	if found {
		panic("An type with that name already exists")
	}

	types[name] = typ
}

// GetType will return a type previously registered with RegisterType.
func GetType(name string) interface{} {
	name = strings.TrimPrefix(name, "*")

	return reflect.New(types[name]).Interface()
}

// NewDatabase will instantiate a new database placed in filepath.
func NewDatabase(filepath string) (*Database, error) {
	d := &Database{}

	err := d.open(path.Join(filepath, "gansoi.db"))
	if err != nil {
		return nil, err
	}

	return d, nil
}

func (d *Database) open(filepath string) error {
	db, err := storm.Open(
		filepath,
		storm.BoltOptions(0600, &bolt.Options{Timeout: 1 * time.Second}),
		storm.AutoIncrement(),
	)
	if err != nil {
		return err
	}

	d.db = db

	return nil
}

// NewLogEntry will return a new LogEntry ready for committing to the Raft log.
func NewLogEntry(cmd Command, value interface{}) *LogEntry {
	v, _ := json.Marshal(value)

	return &LogEntry{
		Command: cmd,
		Type:    reflect.TypeOf(value).String(),
		Value:   v,
	}
}

// Byte is a simple helper, that will marshal the entry to a byte slice.
func (e *LogEntry) Byte() []byte {
	b, _ := json.Marshal(e)

	return b
}

// Storm will return the underlying Storm database.
func (d *Database) Storm() *storm.DB {
	d.dbMutex.RLock()
	defer d.dbMutex.RUnlock()

	return d.db
}

// ProcessLogEntry will process the log entry and apply whatever needs doing.
func (d *Database) ProcessLogEntry(entry *LogEntry) error {
	var err error

	switch entry.Command {
	case CommandSave:
		v := GetType(entry.Type)
		json.Unmarshal(entry.Value, v)
		d.Save(v)
	case CommandDelete:
		v := GetType(entry.Type)
		json.Unmarshal(entry.Value, v)
		err = d.db.DeleteStruct(v)
	default:
		err = fmt.Errorf("not implemented")
	}

	if err != nil {
		panic(err.Error())
	}
	return err
}

// Apply implements raft.FSM.
func (d *Database) Apply(l *raft.Log) interface{} {
	entry := &LogEntry{}
	err := json.Unmarshal(l.Data, entry)
	if err != nil {
		// This should not happen..?
		panic(err.Error())
	}

	return d.ProcessLogEntry(entry)
}

// Snapshot implements raft.FSM.
func (d *Database) Snapshot() (raft.FSMSnapshot, error) {
	return &Snapshot{db: d}, nil
}

// Restore implements raft.FSM.
func (d *Database) Restore(source io.ReadCloser) error {
	db := d.Storm().Bolt
	d.dbMutex.Lock()
	defer d.dbMutex.Unlock()
	defer source.Close()

	path := db.Path()
	restorePath := path + ".restoretmp"

	file, err := os.Create(restorePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, source)
	if err != nil {
		return err
	}

	err = db.Close()
	if err != nil {
		return err
	}

	err = os.Remove(path)
	if err != nil {
		return err
	}

	err = os.Rename(restorePath, path)
	if err != nil {
		return err
	}

	err = d.open(path)
	if err != nil {
		return err
	}

	return nil
}

// Save will save an object to the database.
func (d *Database) Save(data interface{}) error {
	d.dbMutex.RLock()
	defer d.dbMutex.RUnlock()

	return d.db.Save(data)
}

// One will retrieve one record from the database.
func (d *Database) One(fieldName string, value interface{}, to interface{}) error {
	d.dbMutex.RLock()
	defer d.dbMutex.RUnlock()

	return d.db.One(fieldName, value, to)
}

// All lists all kinds of a type.
func (d *Database) All(to interface{}, limit int, skip int, reverse bool) error {
	d.dbMutex.RLock()
	defer d.dbMutex.RUnlock()

	return d.db.All(to, func(opts *index.Options) {
		opts.Limit = limit
		opts.Skip = skip
		opts.Reverse = reverse
	})
}
