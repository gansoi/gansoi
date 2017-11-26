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

func (_ *writeFailer) Read(b []byte) (n int, err error) {
	return 0, nil
}

func (_ *writeFailer) Write(b []byte) (n int, err error) {
	return 0, errors.New("write failed")
}

func (_ *writeFailer) Close() error {
	return nil
}

func (_ *writeFailer) LocalAddr() net.Addr {
	return &net.TCPAddr{}
}

func (_ *writeFailer) RemoteAddr() net.Addr {
	return &net.TCPAddr{}
}

func (_ *writeFailer) SetDeadline(t time.Time) error {
	return nil
}

func (_ *writeFailer) SetReadDeadline(t time.Time) error {
	return nil
}

func (_ *writeFailer) SetWriteDeadline(t time.Time) error {
	return nil
}
