package http

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gansoi/gansoi/plugins"
)

func TestAgent(t *testing.T) {
	a := plugins.GetAgent("http")
	_ = a.(*HTTP)
}

func TestGetHostPort(t *testing.T) {
	cases := []struct {
		in   string
		host string
		port string
	}{
		{"http://gansoi.com/", "gansoi.com", "80"},
		{"http://gansoi.com:81/", "gansoi.com", "81"},
		{"https://gansoi.com/", "gansoi.com", "443"},
		{"https://gansoi.com:81/", "gansoi.com", "81"},
	}

	for _, c := range cases {
		URL, err := url.Parse(c.in)
		if err != nil {
			t.Fatalf("Test case '%s' cannot be parsed as an URL: %s", c.in, err.Error())
		}

		host, port := getHostPort(URL)
		if host != c.host {
			t.Fatalf("'%s' did not return expected hostname '%s', got '%s'", c.in, c.host, host)
		}

		if port != c.port {
			t.Fatalf("'%s' did not return expected port '%s', got '%s'", c.in, c.port, port)
		}
	}
}

func TestCheckFail(t *testing.T) {
	cases := []string{
		"",
		"http://",
		"example.com",
		"non://",
		"/",
		"https://127.0.0.1:0/",
		"http://127.0.0.1:0/",
		"http/////",
		"http://go-test-nonexisting/",
	}

	for _, u := range cases {
		a := &HTTP{
			URL: u,
		}

		result := plugins.NewAgentResult()
		err := a.Check(result)
		if err == nil {
			t.Fatalf("Failed to detect error for '%s'", u)
		}
	}
}

func okHandler(w http.ResponseWriter, r *http.Request) {
	if r.RequestURI == "/redirect" {
		http.Redirect(w, r, "/", 301)
		return
	}

	fmt.Fprintf(w, "ok\n")
}

func TestCheck(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(okHandler))
	a := &HTTP{
		URL: ts.URL + "/",
	}

	result := plugins.NewAgentResult()
	err := a.Check(result)
	if err != nil {
		t.Fatalf("Check failed: %s", err.Error())
	}
	ts.Close()

	ts = httptest.NewTLSServer(http.HandlerFunc(okHandler))
	ts.Config.ErrorLog = log.New(ioutil.Discard, "", 0)

	a.URL = ts.URL + "/"
	err = a.Check(result)
	if err == nil {
		t.Fatalf("Check did not catch unsigned cert")
	}

	a.Insecure = true
	a.IncludeBody = true
	err = a.Check(result)
	if err != nil {
		t.Fatalf("Check failed: %s", err.Error())
	}

	if strings.TrimSpace(result["Body"].(string)) != "ok" {
		t.Fatalf("Got wrong body, expected 'ok', got '%s'", result["Body"])
	}

	a.URL = ts.URL + "/redirect"
	a.FollowRedirect = true
	a.Insecure = true
	a.IncludeBody = true
	err = a.Check(result)
	if err != nil {
		t.Fatalf("Check failed: %s", err.Error())
	}

	ts.Close()
}

var _ plugins.Agent = (*HTTP)(nil)
