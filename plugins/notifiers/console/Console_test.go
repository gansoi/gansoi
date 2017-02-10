package console

import (
	"testing"

	"github.com/gansoi/gansoi/plugins"
)

func TestNotifier(t *testing.T) {
	n := plugins.GetNotifier("console")
	_ = n.(*Console)
}

var _ plugins.Notifier = (*Console)(nil)
