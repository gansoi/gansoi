package ping

import (
	"fmt"
	"os"
	"testing"

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

	fmt.Printf("%+v\n", result)
}

func TestMain(m *testing.M) {
	if !Available() {
		fmt.Printf("ICMP not available in this process, skipping some tests.\n")
	}

	os.Exit(m.Run())
}
