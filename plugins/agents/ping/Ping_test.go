package ping

import (
	"testing"
	"time"

	"github.com/gansoi/gansoi/plugins"
)

func TestAgent(t *testing.T) {
	if !Available() {
		t.SkipNow()
	}

	a := plugins.GetAgent("ping")
	_ = a.(*Ping)
}

func TestCheck(t *testing.T) {
	if !Available() {
		t.SkipNow()
	}

	a := Ping{
		Target: "go-test-localhost.gansoi-dev.com",
		Count:  5,
	}

	result := plugins.NewAgentResult()
	err := a.Check(result)
	if err != nil {
		t.Fatalf("Check() failed: %s", err.Error())
	}
}

func TestCheckFaker(t *testing.T) {
	waitForReply = time.Millisecond

	listenPacket = newListener(nil, nil)
	defer func() { listenPacket = listen }()
	saved := available
	available = true
	defer func() { available = saved }()

	i = NewICMPService()
	i.Start()

	a := Ping{
		Target: "go-test-localhost.gansoi-dev.com",
		Count:  5,
	}
	result := plugins.NewAgentResult()
	err := a.Check(result)
	if err != nil {
		t.Fatalf("Check() failed: %s", err.Error())
	}

	a.Target = "go-test-localhost-v6.gansoi-dev.com"
	err = a.Check(result)
	if err != nil {
		t.Fatalf("Check() failed: %s", err.Error())
	}

	i.Stop()
}

func TestCheckFail(t *testing.T) {
	saved := available
	available = true
	defer func() { available = saved }()

	listenPacket = newListener(nil, nil)
	defer func() { listenPacket = listen }()

	i = NewICMPService()
	i.Start()

	a := Ping{
		Target: "go-test-nonexisting.gansoi-dev.com",
		Count:  1,
	}

	result := plugins.NewAgentResult()
	err := a.Check(result)
	if err == nil {
		t.Fatalf("Check() failed to err on unknown host")
	}
}

func TestCheckTimeOut(t *testing.T) {
	saved := available
	available = true
	defer func() { available = saved }()

	listenPacket = newListener(nil, nil)
	defer func() { listenPacket = listen }()

	i = NewICMPService()
	i.Start()

	a := Ping{
		Target: "127.0.0.2",
		Count:  1,
	}

	waitForReply = time.Millisecond

	result := plugins.NewAgentResult()
	err := a.Check(result)
	if err != nil {
		t.Fatalf("Check() failed: %s", err.Error())
	}

	i.Stop()
}

func TestCheckUnavailable(t *testing.T) {
	saved := available
	available = false
	defer func() { available = saved }()

	a := Ping{
		Target: "go-test-nonexisting.gansoi-dev.com",
		Count:  1,
	}

	result := plugins.NewAgentResult()
	err := a.Check(result)
	if err != ErrICMPServiceUnavailable {
		t.Fatalf("Check() failed to report ICMP unavailable")
	}
}
