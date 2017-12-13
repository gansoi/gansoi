package http

import (
	"errors"
	"net"
	"time"
)

type (
	writeFailer struct{}
)

func dialWriteFailer(_, _ string) (net.Conn, error) {
	return &writeFailer{}, nil
}

func (f *writeFailer) Read(b []byte) (n int, err error) {
	return 0, nil
}

func (f *writeFailer) Write(b []byte) (n int, err error) {
	return 0, errors.New("write failed")
}

func (f *writeFailer) Close() error {
	return nil
}

func (f *writeFailer) LocalAddr() net.Addr {
	return &net.TCPAddr{}
}

func (f *writeFailer) RemoteAddr() net.Addr {
	return &net.TCPAddr{}
}

func (f *writeFailer) SetDeadline(t time.Time) error {
	return nil
}

func (f *writeFailer) SetReadDeadline(t time.Time) error {
	return nil
}

func (f *writeFailer) SetWriteDeadline(t time.Time) error {
	return nil
}
