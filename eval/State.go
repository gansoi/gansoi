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

	// StateDown is a Check with failed conditions.
	StateDown State = iota

	// stateMax can be used like 'if state >= stateMax' to check for a valid
	// state.
	stateMax State = iota
)

const (
	red   = "\033[31m"
	green = "\033[32m"
	blue  = "\033[34m"
	reset = "\033[0m"
)

var (
	stateToText = map[State]string{
		StateUnknown: "unknown",
		StateUp:      "up",
		StateDown:    "down",
	}

	textToState = map[string]State{
		"":        StateUnknown,
		"unknown": StateUnknown,
		"up":      StateUp,
		"down":    StateDown,
	}

	stateToJSON = map[State]string{
		StateUnknown: `""`,
		StateUp:      `"up"`,
		StateDown:    `"down"`,
	}

	jsonToState = map[string]State{
		`""`:        StateUnknown,
		`"unknown"`: StateUnknown,
		`"up"`:      StateUp,
		`"down"`:    StateDown,
	}

	stateToHuman = map[State]string{
		StateUnknown: "Unknown",
		StateUp:      "Up",
		StateDown:    "Down",
	}

	stateToColor = map[State]string{
		StateUnknown: blue,
		StateUp:      green,
		StateDown:    red,
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

// MarshalText implements encoding.TextMarshaler.
func (s State) MarshalText() ([]byte, error) {
	name, found := stateToText[s]
	if !found {
		name = stateToText[StateUnknown]
	}

	return []byte(name), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (s *State) UnmarshalText(text []byte) error {
	state, found := textToState[string(text)]
	if !found {
		return fmt.Errorf("cannot unmarshal %s to state", string(text))
	}

	*s = state

	return nil
}
