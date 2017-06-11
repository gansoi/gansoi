package eval

import (
	"fmt"
)

type (
	// State denotes the current state of a Check.
	State int
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

const (
	red    = "\033[31m"
	yellow = "\033[33m"
	green  = "\033[32m"
	blue   = "\033[34m"
	reset  = "\033[0m"
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

	stateToHuman = map[State]string{
		StateUnknown:  "Unknown",
		StateUp:       "Up",
		StateDegraded: "Degraded",
		StateDown:     "Down",
	}

	stateToColor = map[State]string{
		StateUnknown:  blue,
		StateUp:       green,
		StateDegraded: yellow,
		StateDown:     red,
	}
)

// String implements GoStringer.
func (s State) String() string {
	return stateToHuman[s]
}

// ColorString will return a colorized string representing the state. It will
// be suitable for printing in an ANSI terminal.
func (s State) ColorString() string {
	return stateToColor[s] + s.String() + reset
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
