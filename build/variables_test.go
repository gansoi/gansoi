package build

import (
	"testing"
)

// We need this to run before init. This is ugly.
var _ = func() (_ struct{}) {
	SHA = "0444c0a26e1a4052902605c0fe997de94ef30f4d"
	timestamp = "1507558055"
	return
}()

func TestTimestamp(t *testing.T) {
	expected := "2017-10-09T14:07:35Z"
	if Time != expected {
		t.Errorf("Time does not contain expected value, got %s expected %s", Time, expected)
	}
}

func TestShortSHA(t *testing.T) {
	if ShortSHA != "0444c0a" {
		t.Errorf("ShortSHA does not contain expected hash")
	}
}
