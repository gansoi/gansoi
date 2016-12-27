package checks

import (
	"time"

	"github.com/gansoi/gansoi/plugins"
)

type (
	// CheckResult describes the result of one or more checks after a single
	// node has executed the check.
	CheckResult struct {
		ID        int64               `json:"id,omitempty"`
		CheckID   string              `json:"check_id,omitempty"`
		Node      string              `json:"node_id,omitempty"`
		Error     string              `json:"error"`
		TimeStamp time.Time           `json:"timestamp"`
		Results   plugins.AgentResult `json:"results"`
	}
)
