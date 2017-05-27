package mock

import (
	"errors"
	"io"
	"net"
)

type (
	// Mock is a transport for testing agents.
	Mock struct{}
)

var (
	// ErrNotImplemented will be returned for methods not implemented.
	ErrNotImplemented = errors.New("not implemented")
)

// Dial should mimic net.Dial.
func (l *Mock) Dial(network, address string) (net.Conn, error) {
	return nil, ErrNotImplemented
}

// Exec should execute a binary on the host.
func (l *Mock) Exec(cmd string, arguments ...string) (io.Reader, io.Reader, error) {
	return nil, nil, ErrNotImplemented
}

// ReadFile should mimic ioutil.ReadFile.
func (l *Mock) ReadFile(filename string) ([]byte, error) {
	return nil, ErrNotImplemented
}
