package checks

import (
	"testing"
	"time"
)

func TestMeta(t *testing.T) {
	c := &Check{Interval: time.Second * 30}
	c.ID = "hello"

	m1 := meta(time.Now(), c)
	if m1 == nil {
		t.Fatalf("meta() failed to return meta")
	}

	m2 := meta(time.Now(), c)
	if m2 == nil {
		t.Fatalf("meta() failed to return meta")
	}

	if m1 != m2 {
		t.Fatalf("meta() failed to return the same meta for the same check")
	}
}
