package database

import "time"

type (
	// CheckResult describes the result of one or more checks.
	CheckResult struct {
		CheckID   string      `json:"check_id"`
		Node      string      `json:"node_id"`
		Error     string      `json:"error"`
		TimeStamp time.Time   `json:"timestamp"`
		Results   interface{} `json:"results"`
	}
)

func init() {
	RegisterType(CheckResult{})
}
