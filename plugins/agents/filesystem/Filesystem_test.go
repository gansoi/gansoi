package filesystem

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
		stdout         io.Reader
		stderr         io.Reader
		transportError error
	}
	ReaderMock struct{}
	ParserMock struct {
		parseError error
	}
)

func (m *TransportMock) Exec(cmd string, arguments ...string) (io.Reader, io.Reader, error) {
	return m.stdout, m.stderr, m.transportError
}

func (m *ReaderMock) Read(p []byte) (int, error) {
	return 0, errors.New("I pretend, that I could not read command's stdout")
}

func (m *ParserMock) parse(p []byte) ([]filesystemInfo, error) {
	return []filesystemInfo{filesystemInfo{Device: "/dev/sda1"}}, m.parseError
}

func TestRemoteCheck(t *testing.T) {
	transport := &TransportMock{stdout: bytes.NewBufferString("test")}
	fs := &Filesystem{}
	fs.parser = &ParserMock{}
	checkError := fs.RemoteCheck(transport, plugins.NewAgentResult())
	if checkError != nil {
		t.Error("RemoteCheck should not fail in this case")
	}
}

func TestRemoteCheckFailsOnParserFailed(t *testing.T) {
	transport := &TransportMock{stdout: bytes.NewBufferString("test")}
	fs := &Filesystem{}
	fs.parser = &ParserMock{parseError: errors.New("parser error")}
	checkError := fs.RemoteCheck(transport, plugins.NewAgentResult())
	if checkError == nil {
		t.Error("RemoteCheck should fail if parser.parse failed")
	}
}

func TestRemoteCheckFailsOnInvokeRemoteCommandFailed(t *testing.T) {
	transport := &TransportMock{stdout: nil, stderr: nil, transportError: errors.New("test")}
	fs := &Filesystem{}
	checkError := fs.RemoteCheck(transport, plugins.NewAgentResult())
	if checkError == nil {
		t.Error("RemoteCheck should fail if invokeRemoteCommand failed")
	}
}

func TestInvokeRemoteCommand(t *testing.T) {
	transport := &TransportMock{stdout: bytes.NewBufferString("test")}
	fs := &Filesystem{}
	response, invokeError := fs.invokeRemoteCommand(transport)
	if invokeError != nil {
		t.Error("invokeError should be nil")
	}
	if bytes.Compare(response, []byte("test")) != 0 {
		t.Error("Invalid response, should be the same as the one passed to mock")
	}
}
func TestInvokeRemoteCommandTransportError(t *testing.T) {
	transport := &TransportMock{stdout: nil, stderr: nil, transportError: errors.New("test")}
	fs := &Filesystem{}
	response, invokeError := fs.invokeRemoteCommand(transport)
	if response != nil {
		t.Error("response should be nil")
	}
	if invokeError == nil {
		t.Error("Transport failed, therefore invokeError should also fail")
	}
}
func TestInvokeRemoteCommandReadError(t *testing.T) {
	transport := &TransportMock{stdout: &ReaderMock{}, stderr: nil, transportError: nil}
	fs := &Filesystem{}
	_, invokeError := fs.invokeRemoteCommand(transport)
	if invokeError == nil {
		t.Error("There was a read error, therefore invokeError should also fail")
	}
}

func TestSetResult(t *testing.T) {
	fsInfos := []filesystemInfo{
		filesystemInfo{Device: "/dev/sda1", Mountpoint: "/", Total: 10, Availabe: 5, Used: 5, UsedPercent: 50},
		filesystemInfo{Device: "/dev/sda2", Mountpoint: "/tmp", Total: 10, Availabe: 1, Used: 9, UsedPercent: 90},
	}
	fs := Filesystem{}
	result := plugins.NewAgentResult()
	fs.setResult(result, fsInfos)
	if result["RootUsed"].(int64) != 5 {
		t.Error("Root device's stats were not included")
	}
	if result["WorstUsed"].(int64) != 9 {
		t.Error("Worst device's stats were not included")
	}
}

func TestSetResultNoRoot(t *testing.T) {
	fsInfos := []filesystemInfo{
		filesystemInfo{Device: "/dev/sda1", Mountpoint: "/foo", Total: 10, Availabe: 5, Used: 5, UsedPercent: 50},
	}
	fs := Filesystem{}
	result := plugins.NewAgentResult()
	fs.setResult(result, fsInfos)
	_, ok := result["RootUsed"]
	if ok {
		t.Error("Root device's stats should not be included")
	}
	if result["WorstDevice"].(string) != "/dev/sda1" {
		t.Error("Worst device's stats were not included")
	}
}
func TestSetResultEmpty(t *testing.T) {
	fsInfos := []filesystemInfo{}
	fs := Filesystem{}
	result := plugins.NewAgentResult()
	err := fs.setResult(result, fsInfos)
	if err == nil {
		t.Error("The check should fail if there are no filesystems")
	}
}
func TestSetResultWithExcludes(t *testing.T) {
	fsInfos := []filesystemInfo{filesystemInfo{Device: "/dev/sda1"}}
	fs := Filesystem{CommaSeparatedExcludedDevices: "/dev/sda1,/dev/sda2"}
	result := plugins.NewAgentResult()
	err := fs.setResult(result, fsInfos)
	if err == nil {
		t.Error("The mock should be filtered out leaving empty result, that leads to error")
	}
}

func TestFilesystemInfo(t *testing.T) {
	fi := &filesystemInfo{Mountpoint: "/"}
	if !fi.isRoot() {
		t.Fatal("This is a root mountpoint")
	}
	fi.Mountpoint = "/home"
	if fi.isRoot() {
		t.Fatal("This is not a root mountpoint")
	}
}
