package plugins

import (
	"testing"
)

func TestResultAddValue(t *testing.T) {
	r := NewAgentResult()

	_, ok := r["key"]
	if ok {
		t.Fatalf("Non-existing value found")
	}

	r.AddValue("key", "value")
	_, ok = r["key"]
	if !ok {
		t.Fatalf("Value not found")
	}

	if r["key"] != "value" {
		t.Fatalf("Wrong value stored, got %v, expected %v", r["key"], "value")
	}
}

func TestValidateResultKeyRune(t *testing.T) {
	cases := []struct {
		in       rune
		expected bool
	}{
		{'a', true},
		{'A', true},
		{'1', true},
		{'Æ', true},
		{'!', false},
		{'?', false},
		{'&', false},
		{' ', false},
		{'<', false},
		{'=', false},
	}

	for _, c := range cases {
		result := ValidateResultKeyRune(c.in)

		if result != c.expected {
			t.Fatalf("ValidateResultKeyRune(%v) returned %v, expected %v", c.in, result, c.expected)
		}
	}
}
