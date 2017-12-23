package database

import (
	"encoding/json"
	"reflect"
	"strings"
	"sync"
)

type (
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

// RegisterType will register the type with the raft log marshaller.
func RegisterType(v interface{}) {
	typesLock.Lock()
	defer typesLock.Unlock()

	typ := reflect.TypeOf(v)
	name := typ.String()

	_, found := types[name]
	if found {
		// This should only be triggered at startup, and a panic is okay for
		// now.
		panic("An type with that name already exists")
	}

	types[name] = typ
}

// getType will return a type previously registered with RegisterType.
func getType(name string) interface{} {
	name = strings.TrimPrefix(name, "*")

	typ, found := types[name]
	if !found {
		// panic() is okay here. We have nothing better to do.
		panic("Type " + name + " is not registered")
	}

	return reflect.New(typ).Interface()
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

// Payload will return the payload of a logentry - if any. This could be
// replaced by proper JSON marshal/unmarshal functions.
func (e *LogEntry) Payload() (interface{}, error) {
	v := getType(e.Type)
	err := json.Unmarshal(e.Value, v)

	if err != nil {
		return nil, err
	}

	return v, nil
}
