package checks

import (
	"time"

	"github.com/gansoi/gansoi/plugins"
)

type (
	// CheckResult describes the result of one or more checks after a single
	// node has executed the check.
	CheckResult struct {
		ID          int64               `json:"id,omitempty"`
		CheckHostID string              `json:"check_host_id,omitempty" storm:"index"`
		CheckID     string              `json:"check_id" storm:"index"`
		HostID      string              `json:"host_id"`
		Node        string              `json:"node_id,omitempty"`
		Error       string              `json:"error"`
		TimeStamp   time.Time           `json:"timestamp"`
		Results     plugins.AgentResult `json:"results"`
	}
)

// CheckHostID returns a compound key constisting of a check id and a host id.
func CheckHostID(checkID string, hostID string) string {
	return checkID + "::" + hostID
}
