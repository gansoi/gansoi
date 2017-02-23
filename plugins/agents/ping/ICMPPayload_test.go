package ping

import (
	"crypto/md5"
	"testing"
)

func TestPayloadBytes(t *testing.T) {
	p := NewICMPPayload()
	b := p.Bytes()

	if len(b) < 17 {
		t.Fatalf("Encoded payload too short")
	}
}

func TestPayloadRead(t *testing.T) {
	p := NewICMPPayload()
	b := p.Bytes()

	p2 := &ICMPPayload{}
	err := p2.Read(b)
	if err != nil {
		t.Fatalf("Read() failed: %s [%s]", err.Error(), b)
	}

	if !p.Timestamp.Equal(p2.Timestamp) {
		t.Fatalf("Failed to encode/decode timestamp, expected '%s' (%d), got '%s' (%d)", p.Timestamp, p.Timestamp.Nanosecond(), p2.Timestamp, p2.Timestamp.Nanosecond())
	}

	if p.Helo != p2.Helo {
		t.Fatalf("Failed to encode/decode helo, expected '%s', got '%s'", p.Helo, p2.Helo)
	}
}

func TestPayloadReadFail(t *testing.T) {
	p := &ICMPPayload{}

	err := p.Read(nil)
	if err == nil {
		t.Fatalf("Read() failed to catch nil input")
	}

	err = p.Read([]byte("too short"))
	if err == nil {
		t.Fatalf("Read() failed to catch broken input")
	}

	err = p.Read([]byte("this has a broken signature"))
	if err == nil {
		t.Fatalf("Read() failed to catch broken signature")
	}
}

func TestPayloadBrokenJSON(t *testing.T) {
	msg := []byte("broken json")
	h := md5.New()
	h.Write(secret)
	h.Write(msg)
	digest := h.Sum(nil)[:]

	msg = append(digest, msg...)

	p := &ICMPPayload{}
	err := p.Read(msg)
	if err == nil {
		t.Fatalf("Read() failed to catch broken JSON")
	}
}
