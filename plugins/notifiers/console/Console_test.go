package console

import (
	"os"
	"strings"
	"testing"

	"github.com/gansoi/gansoi/plugins"
)

func TestNotifier(t *testing.T) {
	n := plugins.GetNotifier("console")
	_ = n.(*Console)
}

func TestNotify(t *testing.T) {
	real := os.Stdout
	r, w, _ := os.Pipe()

	c := plugins.GetNotifier("console")

	os.Stdout = w
	c.Notify("hellohello")
	os.Stdout = real

	buf := make([]byte, 1024)
	n, _ := r.Read(buf)

	w.Close()

	if !strings.Contains(string(buf[:n]), "hellohello") {
		t.Fatalf("Output did nto contain our log entry")
	}
}

var _ plugins.Notifier = (*Console)(nil)
