package router

import (
	"testing"
)

func TestGenerateRegex(t *testing.T) {
	cases := map[string]string{
		"/":             "^/$",
		"/first":        "^/first$",
		"/first/second": "^/first/second$",
		"/{id}":         "^/(?P<id>[[:alnum:]]*)$",
		"/{id}/more":    "^/(?P<id>[[:alnum:]]*)/more$",
		"/{id}/{id2}":   "^/(?P<id>[[:alnum:]]*)/(?P<id2>[[:alnum:]]*)$",
		"":              "^$",
		"first":         "^first$",
		"first/second":  "^first/second$",
		"{id}":          "^(?P<id>[[:alnum:]]*)$",
		"{id}/more":     "^(?P<id>[[:alnum:]]*)/more$",
		"{id}/{id2}":    "^(?P<id>[[:alnum:]]*)/(?P<id2>[[:alnum:]]*)$",
	}

	for input, expected := range cases {
		result, _ := generateRegex(input)
		if result != expected {
			t.Errorf("Expected '%s' from '%s', got '%s'", expected, input, result)
		}
	}
}
