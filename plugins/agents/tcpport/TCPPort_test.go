package tcpport

import (
	"net"
	"testing"

	"github.com/gansoi/gansoi/plugins"
)

func TestAgent(t *testing.T) {
	a := plugins.GetAgent("tcpport")
	_ = a.(*TCPPort)
}

func TestCheckFail(t *testing.T) {
	a := TCPPort{
		Address: "127.0.0.1:0",
	}

	result := plugins.NewAgentResult()
	err := a.Check(result)
	if err == nil {
		t.Fatalf("Failed to detect error")
	}
}

func TestCheckV4(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")

	a := TCPPort{
		Address: l.Addr().String(),
	}

	result := plugins.NewAgentResult()
	err = a.Check(result)
	if err != nil {
		t.Fatalf("Check failed: %s", err.Error())
	}

	l.Close()
}

func TestCheckV6(t *testing.T) {
	l, err := net.Listen("tcp6", "[::1]:0")

	a := TCPPort{
		Address: l.Addr().String(),
	}

	result := plugins.NewAgentResult()
	err = a.Check(result)
	if err != nil {
		t.Fatalf("Check failed: %s", err.Error())
	}

	l.Close()
}

var _ plugins.Agent = (*TCPPort)(nil)
