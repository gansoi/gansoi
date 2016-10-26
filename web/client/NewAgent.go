package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/abrander/gansoi/checks"
	"github.com/abrander/gansoi/plugins"
	"github.com/abrander/gansoi/web/client/router"
	"github.com/abrander/gansoi/web/client/template"
)

type (
	// NewAgent is a controller for adding a new agent.
	NewAgent struct {
		ID          string
		Interval    int
		Description *plugins.PluginDescription
		Result      *checks.CheckResult
		Arguments   map[string]interface{}
		Tests       []string

		render func() error
	}
)

// NewNewAgent will instantiate a NewAgent-controller.
func NewNewAgent(description *plugins.PluginDescription) (*NewAgent, error) {
	newAgent := &NewAgent{
		Interval:    60,
		Description: description,
		Arguments:   make(map[string]interface{}),
	}

	return newAgent, nil
}

// RenderFunc implements template.Renderer.
func (n *NewAgent) RenderFunc(render func() error) {
	n.render = render
}

// Submit implements template.Submitter
func (n *NewAgent) Submit(button *template.Button, arguments map[string]interface{}) {
	interval := arguments["Interval"].(float64)
	if interval <= 0.0 {
		interval = 30
	}

	n.ID = arguments["ID"].(string)

	delete(arguments, "Interval")
	delete(arguments, "ID")

	check := &Check{
		ID:          n.ID,
		AgentID:     n.Description.Name,
		Interval:    time.Duration(interval) * time.Second,
		Node:        "all",
		Arguments:   arguments,
		Expressions: n.Tests,
	}

	b, _ := json.Marshal(check)

	switch button.Name() {
	case "test":
		resp, _ := http.Post("/test", "application/json", bytes.NewBuffer(b))
		body, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()

		result := checks.CheckResult{}
		err := json.Unmarshal(body, &result)
		if err != nil {
			fmt.Printf("ERROR: %s\n", err.Error())
		}

		n.Result = &result
		n.Arguments = arguments
		err = n.render()
		if err != nil {
			fmt.Printf("Error: %s\n", err.Error())
		}

		// Re-enable the test button.
		button.Enable()

	case "addtest":
		newtest, ok := arguments["NewTest"].(string)
		if ok && newtest != "" {
			n.Tests = append(n.Tests, arguments["NewTest"].(string))

			n.render()
		}

		button.Enable()

	case "removetest":
		i, _ := strconv.ParseInt(button.Index(), 10, 32)
		n.Tests = append(n.Tests[:i], n.Tests[i+1:]...)

		n.render()
		button.Enable()

	case "add":
		resp, _ := http.Post("/checks", "application/json", bytes.NewBuffer(b))
		resp.Body.Close()

		router.Set("checks" + "/" + n.ID)

		button.Enable()
	}
}
