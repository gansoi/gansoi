package database

import (
	"encoding/json"
	"reflect"
	"strings"
	"sync"
)

type (
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
