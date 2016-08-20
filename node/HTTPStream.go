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
}

// NewHTTPStream will instantiate a new HTTPStream.
func NewHTTPStream(addr string) (*HTTPStream, error) {
	a, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, err
	}

	h := &HTTPStream{
		addr:     a,
		accepted: make(chan net.Conn),
	}

	return h, nil
}

// Dial will dial a remote http endpoint (and implement raft.StreamLayer).
func (h *HTTPStream) Dial(address string, timeout time.Duration) (net.Conn, error) {
	conn, err := net.DialTimeout("tcp", address, timeout)
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
