package checks

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Knetic/govaluate"
	"gopkg.in/go-playground/validator.v9"

	"github.com/gansoi/gansoi/database"
	"github.com/gansoi/gansoi/plugins"
)

type (
	// Check defines a check to be conducted by Gansoi.
	Check struct {
		database.Object `storm:"inline"`
		Name            string          `json:"name" validate:"required"`
		AgentID         string          `json:"agent" validate:"required"`
		Interval        time.Duration   `json:"interval"`
		Arguments       json.RawMessage `json:"arguments"`
		Agent           plugins.Agent   `json:"-"`
		Expressions     []string        `json:"expressions"`
		ContactGroups   []string        `json:"contactgroups"`
	}

	checkProxy struct {
		database.Object
		Name          string          `json:"name"`
		AgentID       string          `json:"agent"`
		Interval      time.Duration   `json:"interval"`
		Node          string          `json:"node"`
		Arguments     json.RawMessage `json:"arguments"`
		Expressions   []string        `json:"expressions"`
		ContactGroups []string        `json:"contactgroups"`
	}
)

// All returns a slice of all checks in db
func All(db database.Database) ([]Check, error) {
	var allChecks []Check
	err := db.All(&allChecks, -1, 0, false)

	return allChecks, err
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *Check) UnmarshalJSON(data []byte) error {
	proxy := checkProxy{}

	err := json.Unmarshal(data, &proxy)
	if err != nil {
		return err
	}

	c.ID = proxy.ID
	c.Name = proxy.Name
	c.AgentID = proxy.AgentID
	c.Interval = proxy.Interval
	c.Arguments = proxy.Arguments
	c.Expressions = proxy.Expressions
	c.ContactGroups = proxy.ContactGroups

	c.Agent = plugins.GetAgent(c.AgentID)
	if c.Agent == nil {
		return errors.New("Agent not found")
	}

	return json.Unmarshal(c.Arguments, &c.Agent)
}

// RunCheck will run a check and return a CheckResult.
func RunCheck(check *Check) (checkResult *CheckResult) {
	agentResult := plugins.NewAgentResult()
	checkResult = &CheckResult{
		CheckID:   check.ID,
		TimeStamp: time.Now(),
		Results:   agentResult,
	}

	defer func() {
		err := recover()

		if err != nil {
			// We don't know the type of 'err', so we let fmt deal with it :)
			checkResult.Error = fmt.Sprintf("%s", err)
		}
	}()

	e := check.Agent.Check(agentResult)

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
			return err
		}

		result, err := e.Evaluate(result)
		if err != nil {
			return err
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

// Validate implements database.Validator.
func (c *Check) Validate(db database.Database) error {
	v := validator.New()
	return v.Struct(c)
}
