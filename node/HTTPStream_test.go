package node

import (
	"bufio"
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

	s := bufio.NewScanner(conn)
	for s.Scan() {
		_, err = conn.Write(s.Bytes())
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

func exchange(conn net.Conn, payload string) string {
	n, err := conn.Write([]byte(payload + "\n"))
	if err != nil {
		panic(err.Error())
	}
	if n != len(payload)+1 {
		panic("fail")
	}

	b := make([]byte, 1024)
	for {
		// Busy waiting.
		n, _ = conn.Read(b)
		if n > 0 {
			break
		}
	}

	return string(b[:n])
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

	time.Sleep(time.Millisecond * 20) // Give echoServer some time to accept.

	payload := "Hello 123"
	result := exchange(conn, payload)

	// This is because the stupid workaround in HTTPStream.Dial(). Sigh.
	if result[0] == 0 {
		result = result[1:]
	}

	if payload != result {
		t.Errorf("Got wrong reply, got '%s', expected '%s'", result, payload)
	}

	payload = "Hello hello"
	result = exchange(conn, payload)

	if payload != result {
		t.Errorf("Got wrong reply, got '%s', expected '%s'", result, payload)
	}
}

var _ raft.StreamLayer = (*HTTPStream)(nil)
