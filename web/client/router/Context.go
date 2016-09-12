package router

import (
	"encoding/json"
	"fmt"
	"html"
	"regexp"

	"github.com/abrander/gansoi/web/client/browser"
)

type (
	Context struct {
		router  *Router
		aborted bool
		params  map[string]string
	}
)

func NewContext(router *Router) *Context {
	return &Context{
		router: router,
		params: make(map[string]string),
	}
}

func generateRegex(raw string) (string, error) {
	regex := "^"

	state := 0
	token := ""

	for i, r := range raw {
		if r == '{' {
			if state != 0 {
				return "", fmt.Errorf("Encountered illegal { at %d", i)
			}

			regex += regexp.QuoteMeta(token)
			state = 1
			token = ""
		} else if r == '}' {
			if state != 1 {
				return "", fmt.Errorf("Encountered illegal } %d", i)
			}

			// Save param to regex
			regex += "(?P<" + token + ">[[:alnum:]]*)"

			state = 0
			token = ""
		} else {
			if state == 0 {
				token += string(r)
			} else if state == 1 {
				token += string(r)
			}
		}
	}

	if state == 1 {
		return "", fmt.Errorf("Missing }")
	}

	regex += regexp.QuoteMeta(token) + "$"

	return regex, nil
}

func (c *Context) Param(name string) string {
	return c.params[name]
}

func (c *Context) HTML(rawHtml string) {
	if c.aborted == true {
		return
	}

	c.router.container.SetInnerHTML(rawHtml)
}

func (c *Context) Text(text string) {
	c.HTML(html.EscapeString(text))
}

func (c *Context) JSON(v interface{}) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return
	}

	c.HTML("<pre>" + html.EscapeString(string(b)) + "</pre>")
}

func (c *Context) Render(renderer browser.Renderer, templateId string, data interface{}) error {
	if c.aborted == true {
		return nil
	}

	return renderer.RenderElement(c.router.container, templateId, data)
}

func (c *Context) Error(err error) {
	if c.aborted == true {
		return
	}

	if err != nil {
		c.Text(err.Error())
		c.aborted = true
	}
}
