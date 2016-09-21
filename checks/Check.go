package checks

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/Knetic/govaluate"

	"github.com/abrander/gansoi/agents"
	"github.com/abrander/gansoi/database"
)

type (
	// Check defines a check to be conducted by Gansoi.
	Check struct {
		ID          string          `json:"id"`
		AgentID     string          `json:"agent"`
		Interval    time.Duration   `json:"interval"`
		Node        string          `json:"node"`
		Arguments   json.RawMessage `json:"arguments"`
		Agent       agents.Agent    `json:"-"`
		Expressions []string        `json:"expressions"`
	}

	checkProxy struct {
		ID          string          `json:"id"`
		AgentID     string          `json:"agent"`
		Interval    time.Duration   `json:"interval"`
		Node        string          `json:"node"`
		Arguments   json.RawMessage `json:"arguments"`
		Expressions []string        `json:"expressions"`
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
	c.Expressions = proxy.Expressions

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

	// If any expressions is defined, we try to evaluate them until one fails.
	if len(check.Expressions) > 0 && e == nil {
		e = check.Evaluate(result)
	}

	if e != nil {
		checkResult.Error = e.Error()
	}

	return checkResult
}

func newStructMap(in interface{}) map[string]interface{} {
	r := make(map[string]interface{})
	v := reflect.ValueOf(in)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// We only accept structs, return an empty map we're served anything else.
	if v.Kind() != reflect.Struct {
		return r
	}

	typ := v.Type()
	for i := 0; i < v.NumField(); i++ {
		fi := typ.Field(i)
		r[fi.Name] = v.Field(i).Interface()
	}

	return r
}

// Evaluate will evaluate the CheckResult based on a slice of expressions.
func (c *Check) Evaluate(results interface{}) error {
	// govaluate expects a map of values.
	m := newStructMap(results)

	for _, exp := range c.Expressions {
		e, err := govaluate.NewEvaluableExpression(exp)
		if err != nil {
			break
		}

		result, err := e.Evaluate(m)
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
