package node

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"expvar"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gansoi/gansoi/ca"
	"github.com/gansoi/gansoi/cluster"
	"github.com/gansoi/gansoi/logger"
	"github.com/hashicorp/raft"
)

// HTTPStream implements a raft stream for use with Golang's net/http.
type HTTPStream struct {
	closed       int32
	addr         string
	accepted     chan net.Conn
	dial         net.Dialer
	certificates []tls.Certificate
	ca           *ca.CA
	rootCAs      *x509.CertPool
}

var (
	dialed   = expvar.NewInt("http_dialed")
	failed   = expvar.NewInt("http_failed")
	served   = expvar.NewInt("http_served")
	accepted = expvar.NewInt("http_accepted")
)

// NewHTTPStream will instantiate a new HTTPStream.
func NewHTTPStream(addr string, certificates []tls.Certificate, coreCA *ca.CA) (*HTTPStream, error) {
	// Set up our own dialer.
	dial := net.Dialer{
		// Some crappy NAT devices will close a connection after 30 seconds of
		// inactivity. We try to keep alive every 25th second to counter this.
		KeepAlive: time.Second * 25,
	}

	h := &HTTPStream{
		addr:         addr,
		accepted:     make(chan net.Conn),
		dial:         dial,
		certificates: certificates,
		ca:           coreCA,
		rootCAs:      coreCA.CertPool(),
	}

	return h, nil
}

// Dial will dial a remote http endpoint (and implement raft.StreamLayer).
func (h *HTTPStream) Dial(address raft.ServerAddress, timeout time.Duration) (net.Conn, error) {
	var conn net.Conn
	var err error

	// Make a copy of our dialer to allow custom timeout.
	dial := h.dial
	dial.Timeout = timeout

	if !strings.ContainsRune(string(address), ':') {
		address += ":4934"
	}

	dialed.Add(1)

	logger.Debug("httpstream", "Dialing %s", address)

	conf := &tls.Config{
		RootCAs:            h.rootCAs,
		Certificates:       h.certificates,
		ServerName:         string(address),
		InsecureSkipVerify: true,
	}
	conn, err = tls.DialWithDialer(&dial, "tcp", string(address), conf)

	if err != nil {
		failed.Add(1)
		fmt.Printf("ERRRRRROR %s\n", err.Error())
		return nil, err
	}

	// We use Upgrade, and hope that will make proxies happy.
	open := fmt.Sprintf("GET %s/raft HTTP/1.1\nHost: %s\nUpgrade: raft-0\n\n", cluster.CorePrefix, address)

	_, err = conn.Write([]byte(open))
	if err != nil {
		conn.Close()

		failed.Add(1)
		return nil, err
	}

	// This is so amazingly stupid. For some reason modern Go often eats
	// exactly one byte of a hijacked HTTP connection. This is our stupid
	// workaround caused by our inability to diagnose the problem properly.
	// We send one null-byte that will be discarded somewhere. We have no idea
	// why this works, but it does. For now. Sigh.
	// (This is probably our own fault)
	conn.Write([]byte{0})

	return conn, nil
}

// Accept waits for and returns the next connection to the listener.
func (h *HTTPStream) Accept() (net.Conn, error) {
	if atomic.LoadInt32(&h.closed) == 1 {
		return nil, errors.New("Server is shutting down")
	}

	accepted.Add(1)

	return <-h.accepted, nil
}

// Close closes the listener.
// Any blocked Accept operations will be unblocked and return errors.
func (h *HTTPStream) Close() error {
	atomic.StoreInt32(&h.closed, 1)

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
	served.Add(1)

	if h.closed == 1 {
		http.Error(w, "Server is shutting down", http.StatusServiceUnavailable)
		return
	}

	upgrade := r.Header.Get("Upgrade")
	if upgrade != "raft-0" {
		http.Error(w, "This endpoint is for streaming raft", http.StatusBadRequest)
		return
	}

	_, err := h.ca.VerifyHTTPRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
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
