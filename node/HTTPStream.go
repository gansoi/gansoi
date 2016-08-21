package node

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"
)

// HTTPStream implements a raft stream for use with Golang's net/http.
type HTTPStream struct {
	closed   bool
	addr     net.Addr
	accepted chan net.Conn
	dial     net.Dialer
}

// NewHTTPStream will instantiate a new HTTPStream.
func NewHTTPStream(addr string) (*HTTPStream, error) {
	a, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, err
	}

	// We try to derive the local address too. This doesn't make much sense
	// in the real world, but it makes debugging networking issues with
	// multiple nodes on the same host much easier.
	// This should never fail as long as the last call to ResolveTCPAddr()
	// with the same input went well.
	local, _ := net.ResolveTCPAddr("tcp", addr)
	local.Port = 0

	// Set up our own dialer.
	dial := net.Dialer{
		LocalAddr: local,

		// Some crappy NAT devices will close a connection after 30 seconds of
		// inactivity. We try to keep alive every 25th second to counter this.
		KeepAlive: time.Second * 25,
	}

	h := &HTTPStream{
		addr:     a,
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

	conn, err := dial.Dial("tcp", address)
	if err != nil {
		return nil, err
	}

	open := fmt.Sprintf("GET /raft HTTP/1.1\nHost: %s\nUpgrade: raft-0\n\n", h.addr.String())

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
	return h.addr
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
