package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	htmltemplate "html/template"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/abrander/gansoi/plugins"
	"github.com/abrander/gansoi/web/client/template"
)

type (
	// NewAgent is a controller for adding a new agent.
	NewAgent struct {
		Description *plugins.PluginDescription
		Parts       []htmltemplate.HTML
	}
)

// NewNewAgent will instantiate a NewAgent-controller.
func NewNewAgent(description *plugins.PluginDescription, templates *template.Collection) (*NewAgent, error) {
	newAgent := &NewAgent{
		Description: description,
	}

	for _, argument := range description.Arguments {
		tmpl, err := templates.RenderString("agent-argument-"+argument.Type, argument)

		if err != nil {
			return nil, err
		}

		newAgent.Parts = append(newAgent.Parts, htmltemplate.HTML(tmpl))
	}

	return newAgent, nil
}

// Submit implements template.Submitter
func (n *NewAgent) Submit(arguments map[string]interface{}) {
	interval := arguments["Interval"].(float64)
	if interval <= 0.0 {
		interval = 30
	}

	id := arguments["ID"].(string)

	delete(arguments, "Interval")
	delete(arguments, "ID")

	check := &Check{
		ID:        id,
		AgentID:   n.Description.Name,
		Interval:  time.Duration(interval) * time.Second,
		Node:      "all",
		Arguments: arguments,
	}

	b, _ := json.Marshal(check)
	resp, _ := http.Post("/checks", "application/json", bytes.NewBuffer(b))
	body, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	fmt.Printf("New check: %s\n", string(body))
}
