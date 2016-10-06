package eval

type (
	// States can hold multiple states.
	States []State
)

// Add a new state to a States object. The newest state will always be available
// at position 0. This type is not meant to be read directly thou.
func (s *States) Add(state State, clamp int) {
	*s = append(States{state}, *s...)

	s.Clamp(clamp)
}

// Clamp removes elements older than size.
func (s *States) Clamp(size int) {
	if size >= 0 && len(*s) > size {
		*s = (*s)[:size]
	}
}

// Histogram generates a map of count of the content.
func (s *States) Histogram() map[State]int {
	ret := make(map[State]int)

	for _, state := range *s {
		ret[state]++
	}

	return ret
}

// Last returns a new States object containing the latest n states.
func (s *States) Last(n int) *States {
	l := len(*s)

	if n <= 0 || l == 0 {
		return &States{}
	}

	if n > l {
		n = l
	}

	ret := (*s)[:n]

	return &ret
}

// Reduce reduces the slice of states to a single state according to a very
// simple algorithm:
// If there's noe states stores, the result is StateUnknown.
// If any of the states are StateUnknown, return StateUnknown.
// If all states are the same, the result is that state.
// If the contained states are mixed, the result are StateDegraded.
func (s *States) Reduce() State {
	l := len(*s)

	if len(*s) == 0 {
		return StateUnknown
	}

	hist := s.Histogram()

	if hist[StateUnknown] > 0 {
		return StateUnknown
	}

	for state, count := range hist {
		if state >= stateMax {
			return StateUnknown
		}

		if count == l {
			return state
		}
	}

	return StateDegraded
}

// ColorString will return a nicely colored array.
func (s *States) ColorString() string {
	ret := "\033[0m["

	for _, state := range *s {
		ret += " " + state.ColorString()
	}

	ret += " ]"

	return ret
}
