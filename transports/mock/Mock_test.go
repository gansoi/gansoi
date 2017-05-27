package mock

import (
	"testing"

	"github.com/gansoi/gansoi/transports"
)

func TestDial(t *testing.T) {
	mock := &Mock{}
	conn, err := mock.Dial("tcp", "10.0.0.1:12877")
	if conn != nil {
		t.Fatalf("Dial did not return nil")
	}

	if err != ErrNotImplemented {
		t.Fatalf("Dial did not return expected error")
	}
}

func TestExec(t *testing.T) {
	mock := &Mock{}
	stdout, stderr, err := mock.Exec("something")
	if stdout != nil {
		t.Fatalf("Exec did not return nil")
	}

	if stderr != nil {
		t.Fatalf("Exec did not return nil")
	}

	if err != ErrNotImplemented {
		t.Fatalf("Exec did not return expected error")
	}
}

func TestReadFile(t *testing.T) {
	mock := &Mock{}
	contents, err := mock.ReadFile("/")
	if contents != nil {
		t.Fatalf("ReadFile did not return nil")
	}

	if err != ErrNotImplemented {
		t.Fatalf("ReadFile did not return expected error")
	}
}

var _ transports.Transport = (*Mock)(nil)
