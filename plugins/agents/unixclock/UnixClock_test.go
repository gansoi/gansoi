package unixclock

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/gansoi/gansoi/plugins"
	"github.com/gansoi/gansoi/transports/mock"
)

type (
	Mock struct {
		mock.Mock
		Contents  []byte
		noContent bool
	}

	failReader struct {
	}
)

var (
	good = []byte(`1499976758
`)

	empty = []byte("")
	bad1  = []byte("1499976758A\n")
	bad2  = []byte("hello: 763443\n")
	bad3  = []byte("%s\n")
)

func (m *Mock) Exec(cmd string, arguments ...string) (io.Reader, io.Reader, error) {
	if m.noContent {
		return &failReader{}, nil, nil
	}

	return bytes.NewBuffer(m.Contents), bytes.NewBufferString(""), nil
}

func (r failReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("failed")
}

func TestCheck(t *testing.T) {
	transport := &Mock{}
	l := &UnixClock{}
	result := plugins.NewAgentResult()

	transport.Contents = []byte(fmt.Sprintf("%d\n", time.Now().Unix()))
	err := l.RemoteCheck(transport, result)

	if err != nil {
		t.Fatalf("RemoteCheck() returned an error: %s", err.Error())
	}

	skew := result["ClockSkew"].(int)
	if skew < 0 || skew > 1 {
		t.Errorf("ClockSkew seem wrong: %d", skew)
	}

	transport.Contents = []byte(fmt.Sprintf("%d\n", time.Now().Unix()-5))
	err = l.RemoteCheck(transport, result)

	if err != nil {
		t.Fatalf("RemoteCheck() returned an error: %s", err.Error())
	}

	skew = result["ClockSkew"].(int)
	if skew < 5 || skew > 7 {
		t.Errorf("ClockSkew seem wrong: %d", skew)
	}
}

func TestCheckSyntaxError(t *testing.T) {
	cases := [][]byte{empty, bad1, bad2, bad3}
	l := &UnixClock{}
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
	l := &UnixClock{}
	result := plugins.NewAgentResult()

	err := l.RemoteCheck(transport, result)
	if err == nil {
		t.Fatalf("RemoteCheck() did not return an error")
	}

	transport2 := &Mock{
		noContent: true,
	}
	err = l.RemoteCheck(transport2, result)
	if err == nil {
		t.Fatalf("RemoteCheck() did not return an error")
	}
}

var _ plugins.RemoteAgent = (*UnixClock)(nil)
