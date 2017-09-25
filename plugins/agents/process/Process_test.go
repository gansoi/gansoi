package process

import (
	"bytes"
	"io"
	"testing"

	"github.com/pkg/errors"

	"github.com/gansoi/gansoi/plugins"
	"github.com/gansoi/gansoi/transports/mock"
)

type (
	TransportMock struct {
		mock.Mock
		stdout   io.Reader
		stderr   io.Reader
		pidofErr error
	}
	ReaderMock struct{}
)

func (m *TransportMock) Exec(cmd string, arguments ...string) (io.Reader, io.Reader, error) {
	return m.stdout, m.stderr, m.pidofErr
}

func (m *ReaderMock) Read(p []byte) (int, error) {
	return 0, errors.New("I pretend, that I could not read command's stdout")
}

func TestCheckFailedPidof(t *testing.T) {
	p := Process{Name: "httpd"}
	result := plugins.NewAgentResult()
	err := p.RemoteCheck(&TransportMock{stdout: nil, stderr: nil, pidofErr: errors.New("Psst, I pretend to be a real error")}, result)
	if err == nil {
		t.Fatalf("RemoteCheck() did not fail despite pidof process failed")
	}
}

func TestCheckFailedReading(t *testing.T) {
	p := Process{Name: "httpd"}
	result := plugins.NewAgentResult()
	err := p.RemoteCheck(&TransportMock{stdout: &ReaderMock{}, stderr: nil, pidofErr: nil}, result)
	if err == nil {
		t.Fatalf("RemoteCheck() did not fail, despite it could not read command's stdout")
	}
}

func TestCheck(t *testing.T) {
	p := Process{Name: "httpd"}
	result := plugins.NewAgentResult()
	err := p.RemoteCheck(&TransportMock{stdout: bytes.NewBufferString("1337 7331")}, result)
	if err != nil {
		t.Fatalf("RemoteCheck() failed")
	}
	if result["Running"].(int) != 2 {
		t.Fatalf("Setting result variables failed")
	}
}
