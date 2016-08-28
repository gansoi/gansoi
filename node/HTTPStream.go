package node

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

// HTTPStream implements a raft stream for use with Golang's net/http.
type HTTPStream struct {
	closed   bool
	addr     string
	accepted chan net.Conn
	dial     net.Dialer
}

// NewHTTPStream will instantiate a new HTTPStream.
func NewHTTPStream(addr string) (*HTTPStream, error) {
	localAddress := addr
	if strings.Index(localAddress, ":") < 0 {
		localAddress += ":0"
	}

	// We try to derive the local address. This doesn't make much sense
	// in the real world, but it makes debugging networking issues with
	// multiple nodes on the same host much easier.
	local, err := net.ResolveTCPAddr("tcp", localAddress)
	if err != nil {
		return nil, err
	}

	local.Port = 0

	// Set up our own dialer.
	dial := net.Dialer{
		LocalAddr: local,

		// Some crappy NAT devices will close a connection after 30 seconds of
		// inactivity. We try to keep alive every 25th second to counter this.
		KeepAlive: time.Second * 25,
	}

	h := &HTTPStream{
		addr:     addr,
		accepted: make(chan net.Conn),
		dial:     dial,
	}

	return h, nil
}

// Dial will dial a remote http endpoint (and implement raft.StreamLayer).
func (h *HTTPStream) Dial(address string, timeout time.Duration) (net.Conn, error) {
	// Make a copy of our dialer to allow custom timeout.
	dial := h.dial
	dial.Timeout = timeout

	if strings.Index(address, ":") < 0 {
		address += ":443"
	}

	conf := &tls.Config{}
	conn, err := tls.DialWithDialer(&dial, "tcp", address, conf)
	if err != nil {
		return nil, err
	}

	open := fmt.Sprintf("GET /raft HTTP/1.1\nHost: %s\nUpgrade: raft-0\n\n", address)

	_, err = conn.Write([]byte(open))
	if err != nil {
		conn.Close()

		return nil, err
	}

	return conn, nil
}

// Accept waits for and returns the next connection to the listener.
func (h *HTTPStream) Accept() (net.Conn, error) {
	if h.closed {
		return nil, errors.New("Server is shutting down")
	}

	return <-h.accepted, nil
}

// Close closes the listener.
// Any blocked Accept operations will be unblocked and return errors.
func (h *HTTPStream) Close() error {
	h.closed = true

	return nil
}

// Addr returns the listener's network address.
func (h *HTTPStream) Addr() net.Addr {
	return h
}

// String implements net.Addr.
func (h *HTTPStream) String() string {
	return h.addr
}

// Network implements net.Addr.
func (h *HTTPStream) Network() string {
	return "tcp"
}

// ServeHTTP implements the http.Handler interface.
func (h *HTTPStream) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.closed {
		http.Error(w, "Server is shutting down", http.StatusServiceUnavailable)
		return
	}

	upgrade := r.Header.Get("Upgrade")
	if upgrade != "raft-0" {
		http.Error(w, "This endpoint is for streaming raft", http.StatusBadRequest)
		return
	}

	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "webserver doesn't support hijacking", http.StatusInternalServerError)
		return
	}

	// We hijack the connection.
	conn, _, err := hj.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Undo deadlines set by webserver.
	conn.SetDeadline(time.Time{})
	conn.SetWriteDeadline(time.Time{})

	h.accepted <- conn
}
