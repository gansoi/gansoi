package slack

import (
	"testing"

	"github.com/gansoi/gansoi/plugins"
)

func TestNotifier(t *testing.T) {
	n := plugins.GetNotifier("slack")
	_ = n.(*Slack)
}

var _ plugins.Notifier = (*Slack)(nil)
