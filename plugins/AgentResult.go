package plugins

type (
	// AgentResult describes the result from an agent.
	AgentResult struct {
		Values map[string]interface{}
	}
)

// NewAgentResult will instanmtiate a new AgentResult ready for passing to an
// agent.
func NewAgentResult() *AgentResult {
	return &AgentResult{
		Values: make(map[string]interface{}),
	}
}

// AddValue will add a result value.
func (a *AgentResult) AddValue(key string, value interface{}) {
	a.Values[key] = value
}
