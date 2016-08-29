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

	// CheckResult describes the result of one or more checks.
	CheckResult struct {
		ID      int         `json:"id"`
		CheckID string      `json:"check_id"`
		Error   string      `json:"error"`
		First   time.Time   `json:"first"`
		Last    time.Time   `json:"last"`
		Streak  int         `json:"streak"`
		Results interface{} `json:"results"`
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

// Check implements agents.Agent. Will return true if the result changed.
func (c *Check) Check(result *CheckResult) bool {
	result.CheckID = c.ID

	var err error
	var errorString string

	result.Results, err = c.Agent.Check()
	if err != nil {
		errorString = err.Error()
	}

	result.Last = time.Now()
	if result.First == (time.Time{}) {
		result.First = result.Last
	}

	if errorString == result.Error {
		result.Streak++
		return false
	}

	// If we arrive here, we have witnessed a state change. Reset CheckResult
	// and save new error from check.
	result.ID = 0
	result.Streak = 1
	result.First = result.Last
	result.Error = errorString

	return true
}
