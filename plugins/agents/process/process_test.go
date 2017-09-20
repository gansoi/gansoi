package process

import (
	"bytes"
	"io"
	"testing"

	"github.com/gansoi/gansoi/plugins"
	"github.com/gansoi/gansoi/transports/mock"
	"github.com/pkg/errors"
)

type (
	TransportMock struct {
		mock.Mock
		mockedResponse string
	}
)

func (m *TransportMock) Exec(cmd string, arguments ...string) (io.Reader, io.Reader, error) {
	if m.mockedResponse == "" {
		return nil, nil, errors.New("Psst, I pretend to be a real problem")
	}
	return bytes.NewBufferString(m.mockedResponse), bytes.NewBufferString(""), nil
}

func TestCheckFail(t *testing.T) {
	p := Process{Name: "httpd"}
	result := plugins.NewAgentResult()
	err := p.RemoteCheck(&TransportMock{}, result)
	if err == nil {
		t.Fatalf("RemoteCheck() did not fail")
	}
}

func TestCheck(t *testing.T) {
	p := Process{Name: "httpd"}
	result := plugins.NewAgentResult()
	err := p.RemoteCheck(&TransportMock{mockedResponse: "1337 7331"}, result)
	if err != nil {
		t.Fatalf("RemoteCheck() failed")
	}
	if result["pid_1"].(string) != "1337" || result["pid_2"].(string) != "7331" {
		t.Fatalf("Setting result variables failed")
	}
}
