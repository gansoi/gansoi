package router

import (
	"regexp"

	"honnef.co/go/js/dom"

	"github.com/abrander/gansoi/web/client/browser"
)

type (
	routeDescription struct {
		regex *regexp.Regexp
		f     RouterFunc
	}

	Router struct {
		nowShowing string
		container  dom.Element
		quit       chan bool
		routes     map[string]routeDescription
	}

	RouterFunc func(c *Context)
)

func New(container dom.Element) *Router {
	return &Router{
		container: container,
		routes:    make(map[string]routeDescription),
	}
}

func (r *Router) AddRoute(path string, f RouterFunc) error {
	regex, err := generateRegex(path)
	if err != nil {
		return err
	}

	r.routes[path] = routeDescription{
		regex: regexp.MustCompile(regex),
		f:     f,
	}

	return nil
}

// Set the route to path.
func (r *Router) Set(path string) {
	u := browser.Url()
	u.Fragment = path
	browser.SetUrl(*u)
}

func (r *Router) render(hash string) {
	// Iterate over all routes
	for _, description := range r.routes {
		match := description.regex.FindStringSubmatch(hash)

		// If we got more than zero matches, we're good to go.
		if len(match) > 0 {
			context := NewContext(r)

			for i, name := range description.regex.SubexpNames() {
				if name != "" {
					context.params[name] = match[i]
				}
			}

			description.f(context)
			break
		}
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
