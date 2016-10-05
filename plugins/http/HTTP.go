package http

import (
	"net/http"

	"github.com/abrander/gansoi/plugins"
)

func init() {
	plugins.RegisterAgent("http", HTTP{})
}

// HTTP will request a ressource from a HTTP server.
type HTTP struct {
	URL string `json:"url" description:"The URL to request"`
}

// Check implements plugins.Agent.
func (h *HTTP) Check(result *plugins.AgentResult) error {
	resp, err := http.Get(h.URL)
	if err != nil {
		return err
	}

	b := make([]byte, 1024)
	n, _ := resp.Body.Read(b)
	resp.Body.Close()

	result.AddValue("StatusCode", resp.StatusCode)
	result.AddValue("Body", string(b[:n]))

	return nil
}
