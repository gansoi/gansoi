package eval

import (
	"fmt"
)

type (
	// State denotes the current state of a Check.
	State uint8
)

const (
	// StateUnknown is an initial state for checks which are not yet down or up.
	StateUnknown State = iota

	// StateUp is a Check that is OK or "green".
	StateUp State = iota

	// StateDegraded is a Check that is not completely ok or "yellow".
	StateDegraded State = iota

	// StateDown is a Check with failed conditions.
	StateDown State = iota

	// stateMax can be used like 'if state >= stateMax' to check for a valid
	// state.
	stateMax State = iota
)

var (
	stateToJSON = map[State]string{
		StateUnknown:  `""`,
		StateUp:       `"up"`,
		StateDegraded: `"degraded"`,
		StateDown:     `"down"`,
	}

	jsonToState = map[string]State{
		`""`:         StateUnknown,
		`"up"`:       StateUp,
		`"degraded"`: StateDegraded,
		`"down"`:     StateDown,
	}
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
		return fmt.Sprintf("State-%d", s)
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
		return fmt.Sprintf("State-%d", s)
	}
}

// Valid returns true if s is a valid state.
func (s State) Valid() bool {
	return s < stateMax
}

// MarshalJSON implements json.Marshaler.
func (s State) MarshalJSON() ([]byte, error) {
	name, found := stateToJSON[s]
	if !found {
		name = stateToJSON[StateUnknown]
	}

	return []byte(name), nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (s *State) UnmarshalJSON(b []byte) error {
	state, found := jsonToState[string(b)]
	if !found {
		return fmt.Errorf("%s is not a valid state", string(b))
	}

	*s = state

	return nil
}
