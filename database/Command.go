package database

type (
	// Command is used to denote which operation should be carried out as a
	// result of a Raft commit.
	Command int
)

const (
	// CommandSave will save an object in the local database.
	CommandSave Command = iota

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
		return "n/a"
	}
}
