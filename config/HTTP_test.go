package config

import (
	"os"
	"reflect"
	"testing"
)

func TestHTTPBind(t *testing.T) {
	cases := []struct {
		url  string
		bind string
	}{
		{"http://gansoi.com/", "gansoi.com:80"},
		{"http://gansoi.com:81/", "gansoi.com:81"},
		{"https://gansoi.com/", "gansoi.com:443"},
		{"https://gansoi.com:444/", "gansoi.com:444"},
	}

	for i, c := range cases {
		var h HTTP
		h.LocalURL = c.url
		bind := h.Bind()

		if c.bind != bind {
			t.Fatalf("%d: Bind() for '%s' returned %s, expected %s", i, c.url, bind, c.bind)
		}
	}
}

func TestHTTPTLS(t *testing.T) {
	cases := []struct {
		url string
		tls bool
	}{
		{"http://gansoi.com/", false},
		{"http://gansoi.com:81/", false},
		{"https://gansoi.com/", true},
		{"https://gansoi.com:444/", true},
	}

	for i, c := range cases {
		var h HTTP
		h.LocalURL = c.url
		tls := h.TLS()

		if c.tls != tls {
			t.Fatalf("%d: TLS() for '%s' returned %T, expected %T", i, c.url, tls, c.tls)
		}
	}
}

func TestURLToHostname(t *testing.T) {
	cases := []struct {
		in       string
		hostname string
	}{
		{"", ""},
		{"http://gansoi.com/", "gansoi.com"},
		{"https://gansoi.com/", "gansoi.com"},
		{"https://gansoi.com:443/", "gansoi.com"},
	}

	for i, c := range cases {
		hostname := urlToHost(c.in)

		if hostname != c.hostname {
			t.Fatalf("%d: urlToHost() failed for %s, returned %s, expected %s", i, c.in, hostname, c.hostname)
		}
	}
}

func TestHTTPHostnames(t *testing.T) {
	hostname, _ := os.Hostname()

	cases := []struct {
		in        HTTP
		hostnames []string
	}{
		{HTTP{LocalURL: "%0"}, []string{hostname}},
		{HTTP{LocalURL: "http://gansoi.com/"}, []string{hostname, "gansoi.com"}},
		{HTTP{LocalURL: "http://gansoi.com/", ClusterURL: "http://gansoi2.com/"}, []string{hostname, "gansoi.com", "gansoi2.com"}},
		{HTTP{LocalURL: "http://gansoi.com/", ClusterURL: "http://gansoi.com/"}, []string{hostname, "gansoi.com", "gansoi.com"}},
	}

	for i, c := range cases {
		hostnames := c.in.Hostnames()

		if !reflect.DeepEqual(hostnames, c.hostnames) {
			t.Fatalf("%d: Hostnames() failed for %+v, returned %v, expected %v", i, c.in, hostnames, c.hostnames)
		}
	}
}

func TestHTTPValidate(t *testing.T) {
	cases := []struct {
		in  HTTP
		err bool
	}{
		{HTTP{LocalURL: "%0"}, true},
		{HTTP{LocalURL: "hejhej"}, true},
		{HTTP{LocalURL: "hej://gansoi.com/"}, true},
		{HTTP{LocalURL: "hej://gansoi.com/"}, true},
		{HTTP{LocalURL: "http://gansoi.com/"}, false},
		{HTTP{LocalURL: "http://gansoi.com/", ClusterURL: "%0"}, true},
		{HTTP{LocalURL: "http://gansoi.com/", ClusterURL: "hejhej"}, true},
		{HTTP{LocalURL: "http://gansoi.com/", ClusterURL: "hej://gansoi.com/"}, true},
		{HTTP{LocalURL: "http://gansoi.com/", ClusterURL: "hej://gansoi.com/"}, true},
		{HTTP{LocalURL: "http://gansoi.com/", ClusterURL: "http://gansoi.com/"}, false},
	}

	for i, c := range cases {
		err := c.in.Validate()

		if err == nil && c.err {
			t.Fatalf("%d: Validate() did not catch error for %+v", i, c.in)
		}

		if err != nil && !c.err {
			t.Fatalf("%d: Validate() returned error for %+v: %s", i, c.in, err.Error())
		}
	}
}
