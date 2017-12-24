package database

import (
	"reflect"
	"testing"
)

type (
	mockType struct {
		A int
		B int
	}
)

func mustPanic(t *testing.T) {
	if r := recover(); r == nil {
		t.Errorf("test did not cause a panic")
	}
}

func resetTypes() {
	types = make(map[string]reflect.Type)
}

func TestLogEntryDualReg(t *testing.T) {
	defer mustPanic(t)

	resetTypes()

	RegisterType(mockType{})
	RegisterType(mockType{})
}

func TestLogEntryReg(t *testing.T) {
	resetTypes()

	RegisterType(mockType{})
}

func TestLogEntryGetType(t *testing.T) {
	resetTypes()

	RegisterType(mockType{})
	getType("database.mockType")
}

func TestLogEntryGetTypeFail(t *testing.T) {
	defer mustPanic(t)

	resetTypes()

	getType("nonexisting")
}

func TestLogEntryNew(t *testing.T) {
	resetTypes()
	RegisterType(mockType{})

	m := mockType{}
	NewLogEntry(CommandSave, m)
}

func TestLogEntryByte(t *testing.T) {
	resetTypes()
	RegisterType(mockType{})

	m := mockType{A: 12}
	e := NewLogEntry(CommandSave, m)
	b := e.Byte()

	if string(b) != `{"Command":0,"Type":"database.mockType","Value":{"A":12,"B":0}}` {
		t.Fatalf("failed to marshal LogEntry to byte")
	}
}

func TestLogEntryPayload(t *testing.T) {
	resetTypes()
	RegisterType(mockType{})

	m := mockType{A: 12}
	e := NewLogEntry(CommandSave, m)
	i := e.Payload()

	if *(i.(*mockType)) != m {
		t.Fatalf("LogEntry roundtrip failed")
	}
}

func TestLogEntryPayloadFail(t *testing.T) {
	resetTypes()
	RegisterType(mockType{})

	m := mockType{A: 12}
	e := NewLogEntry(CommandSave, m)

	e.Value = nil

	p := e.Payload()
	if p != nil {
		t.Fatalf("e.Payload() did not catch JSON error")
	}
}
