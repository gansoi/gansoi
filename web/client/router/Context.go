package router

import (
	"encoding/json"
	"html"

	"github.com/abrander/agento-io/client/browser"
)

type (
	Context struct {
		router  *Router
		aborted bool
	}
)

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
