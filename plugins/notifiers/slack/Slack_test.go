package slack

import (
	"testing"

	"github.com/gansoi/gansoi/plugins"
)

func TestNotifier(t *testing.T) {
	n := plugins.GetNotifier("slack")
	_ = n.(*Slack)
}

func TestSlackFail(t *testing.T) {
	n := plugins.GetNotifier("slack")
	err := n.Notify("hello")
	if err == nil {
		t.Fatalf("Notify() did not return an error for empty arguments")
	}
}

func TestSlack(t *testing.T) {
	s := Slack{
		URL: "https://hooks.slack.com/services/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",
	}

	err := s.Notify("TestSlack")
	if err != nil {
		t.Fatalf("Notify() returned an error: %s", err.Error())
	}
}

var _ plugins.Notifier = (*Slack)(nil)
