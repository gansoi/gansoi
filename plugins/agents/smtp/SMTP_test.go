package smtp

import (
	"net"
	"testing"

	"github.com/gansoi/gansoi/plugins"
)

type (
	tcpBannerServer struct {
		banner string
	}
)

const (
	banner = "220 test-server ESMTP"
)

func newBannerServer() net.Listener {
	listener, _ := net.Listen("tcp", "127.0.0.1:0")

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}

			conn.Write([]byte(banner + "\r\n"))
		}
	}()

	return listener
}

func TestDefaultPort(t *testing.T) {
	cases := map[string]string{
		"hello":     "hello:25",
		"hello:25":  "hello:25",
		"hello::25": "hello::25",
	}

	for input, expected := range cases {
		output := defaultPort(input)
		if output != expected {
			t.Fatalf("defaultPort() did not return what we expected, got %s, expected %s", output, expected)
		}
	}
}

func TestAgent(t *testing.T) {
	a := plugins.GetAgent("smtp")
	_ = a.(*SMTP)
}

func TestCheckFail(t *testing.T) {
	a := SMTP{
		Address: "127.0.0.1:0",
	}

	result := plugins.NewAgentResult()
	err := a.Check(result)
	if err == nil {
		t.Fatalf("Failed to detect error")
	}
}

func TestCheck(t *testing.T) {
	serve := newBannerServer()
	defer serve.Close()

	a := SMTP{
		Address: serve.Addr().String(),
	}

	result := plugins.NewAgentResult()
	err := a.Check(result)
	if err != nil {
		t.Fatalf("Check failed: %s", err.Error())
	}
	if result["banner"] != "test-server ESMTP" {
		t.Fatalf("banner mismatch, got '%s', expected '%s'", result["banner"], banner)
	}
}

var _ plugins.Agent = (*SMTP)(nil)
