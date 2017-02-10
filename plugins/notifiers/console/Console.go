package console

import (
	"fmt"

	"github.com/gansoi/gansoi/plugins"
)

type (
	// Console will print notifications to stdout.
	Console struct {
	}
)

func init() {
	plugins.RegisterNotifier("console", Console{})
}

// Notify implement plugins.Notifier
func (c *Console) Notify(text string) error {
	fmt.Printf("\033[46m\033[30mNOTIFICATION: %s\033[0m\n", text)

	return nil
}
