package checks

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Knetic/govaluate"

	"github.com/abrander/gansoi/plugins"
)

type (
	// Check defines a check to be conducted by Gansoi.
	Check struct {
		ID            string          `json:"id"`
		AgentID       string          `json:"agent"`
		Interval      time.Duration   `json:"interval"`
		Arguments     json.RawMessage `json:"arguments"`
		Agent         plugins.Agent   `json:"-"`
		Expressions   []string        `json:"expressions"`
		ContactGroups []string        `json:"contactgroups"`
	}

	checkProxy struct {
		ID            string          `json:"id"`
		AgentID       string          `json:"agent"`
		Interval      time.Duration   `json:"interval"`
		Node          string          `json:"node"`
		Arguments     json.RawMessage `json:"arguments"`
		Expressions   []string        `json:"expressions"`
		ContactGroups []string        `json:"contactgroups"`
	}
)

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
	c.Arguments = proxy.Arguments
	c.Expressions = proxy.Expressions
	c.ContactGroups = proxy.ContactGroups

	c.Agent = plugins.GetAgent(c.AgentID)
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
func RunCheck(check *Check) *CheckResult {
	agentResult := plugins.NewAgentResult()
	e := check.Agent.Check(agentResult)

	checkResult := &CheckResult{
		CheckID:   check.ID,
		TimeStamp: time.Now(),
		Results:   agentResult,
	}

	// If any expressions is defined, we try to evaluate them until one fails.
	if len(check.Expressions) > 0 && e == nil {
		e = check.Evaluate(agentResult)
	}

	if e != nil {
		checkResult.Error = e.Error()
	}

	return checkResult
}

// Evaluate will evaluate the CheckResult based on a slice of expressions.
func (c *Check) Evaluate(result plugins.AgentResult) error {
	for _, exp := range c.Expressions {
		e, err := govaluate.NewEvaluableExpression(exp)
		if err != nil {
			break
		}

		result, err := e.Evaluate(result)
		if err != nil {
			break
		}

		// This is a nifty go trick. If result is NOT a bool, this will assign
		// a default bool (false) value to checkResult.Passed - and not panic.
		passed, _ := result.(bool)

		// On first false evaluation, we skip the rest of the tests. We already
		// failed.
		if !passed {
			return fmt.Errorf("(%s) failed", exp)
		}
	}

	return nil
}
