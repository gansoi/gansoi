package node

import (
	"crypto/subtle"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gansoi/gansoi/logger"
	"github.com/gansoi/gansoi/stats"
)

// HTTPStream implements a raft stream for use with Golang's net/http.
type HTTPStream struct {
	closed   bool
	addr     string
	accepted chan net.Conn
	dial     net.Dialer
	secret   string
}

func init() {
	stats.CounterInit("http_dialed")
	stats.CounterInit("http_failed")
	stats.CounterInit("http_served")
	stats.CounterInit("http_accepted")
}

// NewHTTPStream will instantiate a new HTTPStream.
func NewHTTPStream(addr string, secret string) (*HTTPStream, error) {
	// Set up our own dialer.
	dial := net.Dialer{
		// Some crappy NAT devices will close a connection after 30 seconds of
		// inactivity. We try to keep alive every 25th second to counter this.
		KeepAlive: time.Second * 25,
	}

	h := &HTTPStream{
		addr:     addr,
		accepted: make(chan net.Conn),
		dial:     dial,
		secret:   secret,
	}

	return h, nil
}

// Dial will dial a remote http endpoint (and implement raft.StreamLayer).
func (h *HTTPStream) Dial(address string, timeout time.Duration) (net.Conn, error) {
	var conn net.Conn
	var err error
	var host, port string
	// Make a copy of our dialer to allow custom timeout.
	dial := h.dial
	dial.Timeout = timeout

	URL, err := url.Parse(address)
	if err != nil {
		return nil, err
	}

	if strings.IndexRune(URL.Host, ':') < 0 {
		host = URL.Host
	} else {
		host, port, _ = net.SplitHostPort(URL.Host)
	}

	stats.CounterInc("http_dialed", 1)

	logger.Green("httpstream", "Dialing %s [%s]", host, address)

	switch URL.Scheme {
	case "https":
		if port == "" {
			port = "443"
		}

		conf := &tls.Config{
			ServerName: host,
		}
		conn, err = tls.DialWithDialer(&dial, "tcp", net.JoinHostPort(host, port), conf)

	case "http":
		if port == "" {
			port = "80"
		}

		conn, err = dial.Dial("tcp", net.JoinHostPort(host, port))

	default:
		return nil, errors.New("Unknown scheme: " + URL.Scheme)
	}

	if err != nil {
		stats.CounterInc("http_failed", 1)
		return nil, err
	}

	// We use Upgrade, and hope that will make proxies happy.
	open := fmt.Sprintf("GET /raft HTTP/1.1\nHost: %s\nUpgrade: raft-0\nSecret: %s\n\n", host, h.secret)

	_, err = conn.Write([]byte(open))
	if err != nil {
		conn.Close()

		stats.CounterInc("http_failed", 1)
		return nil, err
	}

	return conn, nil
}

// Accept waits for and returns the next connection to the listener.
func (h *HTTPStream) Accept() (net.Conn, error) {
	if h.closed {
		return nil, errors.New("Server is shutting down")
	}

	stats.CounterInc("http_accepted", 1)

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
	stats.CounterInc("http_served", 1)

	if h.closed {
		http.Error(w, "Server is shutting down", http.StatusServiceUnavailable)
		return
	}

	upgrade := r.Header.Get("Upgrade")
	if upgrade != "raft-0" {
		http.Error(w, "This endpoint is for streaming raft", http.StatusBadRequest)
		return
	}

	secret := r.Header.Get("Secret")
	if subtle.ConstantTimeCompare([]byte(h.secret), []byte(secret)) != 1 {
		http.Error(w, "Wrong secret", http.StatusForbidden)
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
