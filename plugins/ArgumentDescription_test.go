package plugins

import (
	"reflect"
	"sync"
	"testing"
)

type (
	Type struct {
		sync.RWMutex
		Number int    `json:"number" description:"A number"`
		Option string `json:"option" description:"Some option" enum:"one,two,three"`
	}
)

func TestGetArguments(t *testing.T) {
	typ := reflect.TypeOf(Type{})

	args := getArguments(typ)

	if len(args) != 2 {
		t.Fatalf("getArguments() returned wrong number of args")
	}

	if args[0].Name != "number" {
		t.Fatalf("Wrong name for Number")
	}

	if args[1].Name != "option" {
		t.Fatalf("Wrong name for Option")
	}

	if len(args[0].EnumValues) != 0 {
		t.Fatalf("EnumValues for Number is not zero-length")
	}

	if len(args[1].EnumValues) != 3 {
		t.Fatalf("EnumValues for Number is not zero-length")
	}

	if args[0].Description != "A number" {
		t.Fatalf("Wrong description for Number")
	}

	if args[1].Description != "Some option" {
		t.Fatalf("Wrong description for Option")
	}
}
