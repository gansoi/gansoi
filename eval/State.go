package eval

type (
	// State denotes the current state of a Check.
	State string
)

const (
	// StateUnknown is an initial state for checks which are not yet down or up.
	StateUnknown State = ""

	// StateUp is a Check that is OK or "green".
	StateUp State = "up"

	// StateDegraded is a Check that is not completely ok or "yellow".
	StateDegraded State = "degraded"

	// StateDown is a Check with failed conditions.
	StateDown State = "down"
)

// String implements GoStringer.
func (s State) String() string {
	switch s {
	case StateUnknown:
		return "Unknown"
	case StateUp:
		return "Up"
	case StateDegraded:
		return "Degraded"
	case StateDown:
		return "Down"
	default:
		// FIXME: Don't panic.
		panic("unknown state")
	}
}

// ColorString will return a colorized string representing the state. It will
// be suitable for printing in an ANSI terminal.
func (s State) ColorString() string {
	const red = "\033[31m"
	const yellow = "\033[33m"
	const green = "\033[32m"
	const blue = "\033[34m"
	const reset = "\033[0m"

	switch s {
	case StateUnknown:
		return blue + "Unknown" + reset
	case StateUp:
		return green + "Up" + reset
	case StateDegraded:
		return yellow + "Degraded" + reset
	case StateDown:
		return red + "Down" + reset
	default:
		// FIXME: Don't panic.
		panic("unknown state")
	}
}
