package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	htmltemplate "html/template"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/abrander/gansoi/plugins"
	"github.com/abrander/gansoi/web/client/template"
)

type (
	// NewAgent is a controller for adding a new agent.
	NewAgent struct {
		description *plugins.PluginDescription
		Parts       []htmltemplate.HTML
	}
)

// NewNewAgent will instantiate a NewAgent-controller.
func NewNewAgent(description *plugins.PluginDescription, templates *template.Collection) (*NewAgent, error) {
	newAgent := &NewAgent{
		description: description,
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
func (n *NewAgent) Submit(values map[string]string) {
	arguments := make(map[string]interface{})

	// Iterate through arguments and build arguments map.
	for _, buh := range n.description.Arguments {
		value, found := values[buh.Name]

		if found {
			switch buh.Type {
			case "string":
				arguments[buh.Name] = value
			default:
				panic("Please implement " + buh.Type + " in NewAgent.Submit()")
			}
		}
	}

	// FIXME: Deal with empty interval.
	interval, _ := strconv.Atoi(values["Interval"])

	check := &Check{
		ID:        values["ID"],
		AgentID:   n.description.Name,
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
