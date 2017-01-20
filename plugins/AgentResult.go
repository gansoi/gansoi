package plugins

import "unicode"

type (
	// AgentResult describes the result from an agent.
	AgentResult map[string]interface{}
)

// NewAgentResult will instanmtiate a new AgentResult ready for passing to an
// agent.
func NewAgentResult() AgentResult {
	return AgentResult(make(map[string]interface{}))
}

// AddValue will add a result value.
func (a AgentResult) AddValue(key string, value interface{}) {
	a[key] = value
}

// ValidateResultKeyRune returns true for a rune allowed in a AgentResult key.
func ValidateResultKeyRune(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsNumber(r) || r == '_'
}
