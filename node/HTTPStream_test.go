package node

import (
	"bytes"
	"crypto/tls"
	"net"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gansoi/gansoi/cluster"
	"github.com/hashicorp/raft"
)

func getStream(t *testing.T) *HTTPStream {
	core := cluster.NewCore(&cluster.Info{})
	core.Bootstrap()

	certificates, _ := core.Start()
	ca := core.CA()

	stream, err := NewHTTPStream("127.0.0.1:0", certificates, ca)
	if err != nil {
		t.Fatalf("NewHTTPStream() returned an error: %s", err.Error())
		return nil
	}

	return stream
}

func echoServer(stream *HTTPStream) {
	conn, err := stream.Accept()
	if err != nil {
		return
	}

	for {
		b := make([]byte, 1024)
		n, err := conn.Read(b)
		if err != nil {
			return
		}

		_, err = conn.Write(b[:n])
		if err != nil {
			return
		}
	}
}

func TestStreamListen(t *testing.T) {
	stream := getStream(t)
	defer stream.Close()

	server := httptest.NewUnstartedServer(stream)
	server.TLS = &tls.Config{Certificates: stream.certificates}
	server.StartTLS()
	defer server.Close()

	conn, err := stream.Dial(raft.ServerAddress(server.Listener.Addr().String()), time.Millisecond*100)
	if err != nil {
		t.Errorf("Dial() errored: %s\n", err.Error())
	}
	defer conn.Close()
}

func exchange(conn net.Conn, payload []byte) []byte {
	n, err := conn.Write(payload)
	if err != nil {
		panic(err.Error())
	}
	if n != len(payload) {
		panic("fail")
	}

	b := make([]byte, 1024)
	n, _ = conn.Read(b)

	return b[:n]
}

func TestStreamEcho(t *testing.T) {
	stream := getStream(t)
	defer stream.Close()

	server := httptest.NewUnstartedServer(stream)
	server.TLS = &tls.Config{Certificates: stream.certificates, ClientAuth: tls.RequestClientCert}
	server.StartTLS()
	defer server.Close()

	go echoServer(stream)

	conn, err := stream.Dial(raft.ServerAddress(server.Listener.Addr().String()), time.Millisecond*100)
	if err != nil {
		t.Fatalf("Dial() errored: %s\n", err.Error())
	}
	defer conn.Close()

	payload := []byte("Hello 123")
	result := exchange(conn, payload)

	if bytes.Compare(payload, result) != 0 {
		t.Errorf("Got wrong reply, got '%s', expected '%s'", result, payload)
	}

	payload = []byte("Hello hello")
	result = exchange(conn, payload)

	if bytes.Compare(payload, result) != 0 {
		t.Errorf("Got wrong reply, got '%s', expected '%s'", result, payload)
	}
}

var _ raft.StreamLayer = (*HTTPStream)(nil)
