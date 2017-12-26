package config

type (
	// HTTPRedirect is the configuration for a HTTP redirector.
	HTTPRedirect struct {
		Bind   string `json:"bind"`
		Target string `json:"target"`
	}
)
