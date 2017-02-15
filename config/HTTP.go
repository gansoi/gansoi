package config

import (
	"errors"
	"net"
	"net/url"
	"os"
	"strings"
)

type (
	// HTTP configures the HTTP server.
	HTTP struct {
		LocalURL   string `toml:"url"`
		ClusterURL string `toml:"cluster-url"`
		CertPath   string `toml:"cert"`
		KeyPath    string `toml:"key"`
		Login      string `toml:"login"`
		Password   string `toml:"password"`
	}
)

const (
	https = "https"
	http  = "http"
)

// Bind will return a string suitable for binding.
func (h HTTP) Bind() string {
	URL, _ := url.Parse(h.LocalURL)

	if strings.IndexRune(URL.Host, ':') > 0 {
		return URL.Host
	}

	var port string

	switch URL.Scheme {
	case https:
		port = "443"
	case http:
		port = "80"
	}

	return net.JoinHostPort(URL.Host, port)
}

// urlToHost extracts the hostname from an URL.
func urlToHost(rawurl string) string {
	URL, err := url.Parse(rawurl)
	if err != nil {
		return ""
	}

	// No port.
	if strings.IndexRune(URL.Host, ':') < 0 {
		return URL.Host
	}

	local, _, _ := net.SplitHostPort(URL.Host)

	return local
}

// Hostnames returns a list of hostnames exposed.
func (h HTTP) Hostnames() []string {
	hostname, _ := os.Hostname()

	hostnames := []string{hostname}

	hostname = urlToHost(h.LocalURL)
	if hostname != "" {
		hostnames = append(hostnames, hostname)
	}

	hostname = urlToHost(h.ClusterURL)
	if hostname != "" {
		hostnames = append(hostnames, hostname)
	}

	return hostnames
}

// TLS will return true if we should set up TLS.
func (h HTTP) TLS() bool {
	URL, _ := url.Parse(h.LocalURL)

	return URL.Scheme == https
}

// Validate tries to validate the configuration.
func (h HTTP) Validate() error {
	URL, err := url.Parse(h.LocalURL)
	if err != nil {
		return errors.New("url is malformed")
	}

	if URL.Scheme != http && URL.Scheme != https {
		return errors.New(URL.Scheme + " is not supported")
	}

	if h.ClusterURL != "" {
		URL, err = url.Parse(h.ClusterURL)
		if err != nil {
			return errors.New("cluster-url is malformed")
		}

		if URL.Scheme != http && URL.Scheme != https {
			return errors.New(URL.Scheme + " is not supported")
		}
	}

	return nil
}
