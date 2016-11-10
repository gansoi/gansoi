package database

import (
	"github.com/abrander/gansoi/logger"
)

type (
	// Command is used to denote which operation should be carried out as a
	// result of a Raft commit.
	Command int
)

const (
	// CommandSave will save an object in the local database.
	CommandSave = iota

	// CommandDelete will delete an object in the local database.
	CommandDelete
)

// String implements Stringer.
func (c Command) String() string {
	switch c {
	case CommandSave:
		return "save"
	case CommandDelete:
		return "delete"
	default:
		logger.Red("database", "Unknown command type '%d'. Please update Command.String().", int(c))
		return "n/a"
	}
}
