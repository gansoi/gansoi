package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/abrander/gansoi/web/client/browser"
	"github.com/abrander/gansoi/web/client/rest"
	"github.com/abrander/gansoi/web/client/router"
	"github.com/abrander/gansoi/web/client/template"
)

type (
	// Check mimics checks.Check.
	Check struct {
		ID        string        `json:"id"`
		AgentID   string        `json:"agent"`
		Interval  time.Duration `json:"interval"`
		Node      string        `json:"node"`
		Arguments interface{}   `json:"arguments"`
	}

	checkList struct {
		List []Check
	}
)

func (c checkList) DeleteCheck(id string) {
	fmt.Printf("Delete %s\n", id)
}

func (c checkList) EditCheck(id string) {
	fmt.Printf("Edit %s\n", id)
}

func main() {
	browser.WaitForLoad()

	url := browser.Url()

	checks := rest.NewClient(url.RawPath+"/checks/", "")

	templates := template.NewCollection("template")

	r := router.New(browser.ID("main"))
	r.AddRoute("overview", func(c *router.Context) {
		c.Render(templates, "overview", nil)
	})

	r.AddRoute("gansoi", func(c *router.Context) {
		type nodeInfo struct {
			Name    string            `json:"name" storm:"id"`
			Started time.Time         `json:"started"`
			Updated time.Time         `json:"updated"`
			Raft    map[string]string `json:"raft"`
		}

		var nodes []nodeInfo
		resp, err := http.Get("/raft/nodes")
		if err != nil {
			c.Render(templates, "error", err.Error())
			return
		}

		decoder := json.NewDecoder(resp.Body)
		decoder.Decode(&nodes)
		err = c.Render(templates, "gansoi", nodes)
		if err != nil {
			c.Render(templates, "error", err.Error())
			return
		}
	})

	r.AddRoute("checks", func(c *router.Context) {
		var list checkList
		err := checks.List(&list.List)

		if err != nil {
			fmt.Println(err.Error())
			c.Render(templates, "error", err.Error())
			return
		}

		err = c.Render(templates, "checks", list)
		if err != nil {
			c.Render(templates, "error", err.Error())
			return
		}
	})

	r.Run()
}
