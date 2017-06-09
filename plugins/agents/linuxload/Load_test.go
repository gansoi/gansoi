package load

import (
	"testing"

	"github.com/gansoi/gansoi/plugins"
	"github.com/gansoi/gansoi/transports/mock"
)

type (
	Mock struct {
		mock.Mock
		Contents []byte
	}
)

var (
	good = []byte(`2.07 1.49 1.33 4/960 23141
`)

	empty = []byte("")
	bad1  = []byte("hello")
	bad2  = []byte("hello: 763443\nhello2:")
	bad3  = []byte("2.07 1.49 1.33 4/960\n")
)

func (m *Mock) ReadFile(path string) ([]byte, error) {
	return m.Contents, nil
}

func TestCheck(t *testing.T) {
	transport := &Mock{Contents: good}
	l := &Load{}
	result := plugins.NewAgentResult()

	err := l.RemoteCheck(transport, result)

	if err != nil {
		t.Fatalf("RemoteCheck() returned an error: %s", err.Error())
	}

	if result["Running"] != 4 {
		t.Fatalf("RemoteCheck() returned wrong value for \"Running\", got %d, expected %d", result["Running"], 4)
	}
}

func TestCheckSyntaxError(t *testing.T) {
	cases := [][]byte{empty, bad1, bad2, bad3}
	l := &Load{}
	result := plugins.NewAgentResult()

	for i, c := range cases {
		transport := &Mock{Contents: c}
		err := l.RemoteCheck(transport, result)
		if err == nil {
			t.Fatalf("%d RemoteCheck() did not return an error", i)
		}
	}
}

func TestCheckError(t *testing.T) {
	transport := &mock.Mock{}
	l := &Load{}
	result := plugins.NewAgentResult()

	err := l.RemoteCheck(transport, result)

	if err != mock.ErrNotImplemented {
		t.Fatalf("RemoteCheck() did not return error")
	}
}

var _ plugins.RemoteAgent = (*Load)(nil)
