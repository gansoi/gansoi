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
// 1) If there's no states stored, return StateUnknown.
// 2) If any state is StateUnknown, return StateUnknown.
// 3) If no state is StateUnknown, and there is a majority that majority is
//    returned.
// 4) If none of 1-3 is satisfied, StateUnknown will we returned.
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
		if count > l/2 {
			return state
		}
	}

	return StateUnknown
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
