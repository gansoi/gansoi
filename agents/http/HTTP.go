package http

import (
	"net/http"

	"github.com/abrander/gansoi/agents"
)

func init() {
	agents.RegisterAgent("http", HTTP{})
}

// HTTP will request a ressource from a HTTP server.
type HTTP struct {
	URL string `json:"url" description:"The URL to request"`
}

// Result is the result from this test.
type Result struct {
	StatusCode int    `description:"Result code from web server"`
	Body       string `description:"The first 1024 bytes of the response body"`
}

// Check implements agents.Agent.
func (h *HTTP) Check() (interface{}, error) {
	r := &Result{}

	resp, err := http.Get(h.URL)
	if err != nil {
		return nil, err
	}

	r.StatusCode = resp.StatusCode

	b := make([]byte, 1024)
	n, _ := resp.Body.Read(b)
	r.Body = string(b[:n])
	resp.Body.Close()

	return r, nil
}
