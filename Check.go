package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"time"

	"github.com/abrander/gansoi/agents"
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

// NewCheckFromReader will instantiate a new check based on a JSON stream.
func NewCheckFromReader(r io.Reader) (*Check, error) {
	var check Check

	decoder := json.NewDecoder(r)
	err := decoder.Decode(&check)
	if err != nil {
		return nil, err
	}

	check.Agent = agents.GetAgent(check.AgentID)
	if check.Agent == nil {
		return nil, errors.New("Agent not found")
	}

	err = json.Unmarshal(check.Arguments, &check.Agent)
	if err != nil {
		return nil, err
	}

	return &check, nil
}

// NewCheckFromBytes will instantiate a new check from a byte array of JSON.
func NewCheckFromBytes(b []byte) (*Check, error) {
	return NewCheckFromReader(bytes.NewBuffer(b))
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
