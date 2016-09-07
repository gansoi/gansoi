package main

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/abrander/gansoi/agents"
	"github.com/abrander/gansoi/database"
)

type (
	// Check defines a check to be conducted by Gansoi.
	Check struct {
		ID        string          `json:"id"`
		AgentID   string          `json:"agent"`
		Interval  time.Duration   `json:"interval"`
		Node      string          `json:"node"`
		Arguments json.RawMessage `json:"arguments"`
		Agent     agents.Agent    `json:"-"`
	}

	checkProxy struct {
		ID        string          `json:"id"`
		AgentID   string          `json:"agent"`
		Interval  time.Duration   `json:"interval"`
		Node      string          `json:"node"`
		Arguments json.RawMessage `json:"arguments"`
	}
)

func init() {
	database.RegisterType(Check{})
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *Check) UnmarshalJSON(data []byte) error {
	proxy := checkProxy{}

	err := json.Unmarshal(data, &proxy)
	if err != nil {
		return err
	}

	c.ID = proxy.ID
	c.AgentID = proxy.AgentID
	c.Interval = proxy.Interval
	c.Node = proxy.Node
	c.Arguments = proxy.Arguments

	c.Agent = agents.GetAgent(c.AgentID)
	if c.Agent == nil {
		return errors.New("Agent not found")
	}

	err = json.Unmarshal(c.Arguments, &c.Agent)
	if err != nil {
		return err
	}

	return nil
}

// RunCheck will run a check and return a CheckResult.
func RunCheck(check *Check) *database.CheckResult {
	result, e := check.Agent.Check()

	checkResult := &database.CheckResult{
		CheckID:   check.ID,
		TimeStamp: time.Now(),
		Results:   result,
	}

	if e != nil {
		checkResult.Error = e.Error()
	}

	return checkResult
}
