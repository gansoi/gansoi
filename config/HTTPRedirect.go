package config

type (
	// HTTPRedirect is the configuration for a HTTP redirector.
	HTTPRedirect struct {
		Bind   string `toml:"bind"`
		Target string `toml:"target"`
	}
)
