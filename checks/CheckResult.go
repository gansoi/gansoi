package checks

import (
	"time"

	"github.com/abrander/gansoi/agents"
)

type (
	// CheckResult describes the result of one or more checks after a single
	// node has executed the check.
	CheckResult struct {
		ID        int64               `json:"id"`
		CheckID   string              `json:"check_id"`
		Node      string              `json:"node_id"`
		Error     string              `json:"error"`
		TimeStamp time.Time           `json:"timestamp"`
		Results   *agents.AgentResult `json:"results"`
	}
)
