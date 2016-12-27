package notify

import (
	"errors"
	"testing"

	"github.com/gansoi/gansoi/plugins"
)

type (
	MockAgent struct {
		PleaseReturn string
	}
)

const (
	name = "mockagent"
)

func init() {
	plugins.RegisterNotifier(name, MockAgent{})
}

func TestGet(t *testing.T) {
	plugins.GetNotifier(name)
}

func (a *MockAgent) Notify(text string) error {
	if a.PleaseReturn != "" {
		return errors.New(a.PleaseReturn)
	}

	return nil
}

func TestMockNotify(t *testing.T) {
	a := MockAgent{}
	err := a.Notify("some message")
	if err != nil {
		t.Fatalf("MockAgent.Notify() returned an error: %s", err.Error())
	}
}

var _ plugins.Notifier = (*MockAgent)(nil)
