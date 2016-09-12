package router

import (
	"strings"

	"honnef.co/go/js/dom"

	"github.com/abrander/gansoi/web/client/browser"
)

type (
	Router struct {
		nowShowing string
		container  dom.Element
		quit       chan bool
		routes     map[string]RouterFunc
	}

	RouterFunc func(c *Context)
)

func New(container dom.Element) *Router {
	return &Router{
		container: container,
		routes:    make(map[string]RouterFunc),
	}
}

func (r *Router) AddRoute(path string, f RouterFunc) error {
	r.routes[path] = f

	return nil
}

func (r *Router) render(hash string) {
	var f RouterFunc
	matchLen := 0

	// Iterate over all routes
	for key, value := range r.routes {

		// Check if prefix matches
		if strings.HasPrefix(hash, key) {
			l := len(key)

			// When several destinations matches, the route with the
			// longest key is chosen (the most specific one)
			if l > matchLen {
				matchLen = l
				f = value
			}
		}
	}

	// If f is not nil it means that we found a match
	if f != nil {
		context := &Context{
			router: r,
		}

		f(context)
	}
}

// Run router main loop
func (r *Router) Run() error {
	// Render initial view
	r.render(browser.Url().Fragment)

	change := browser.NewHashChangeListener()

	for {
		select {
		case newHash := <-change.C:
			r.render(newHash)
		case <-r.quit:
			change.Close()
			break
		}
	}
}

func (r *Router) Quit() {
	r.quit <- true
}
