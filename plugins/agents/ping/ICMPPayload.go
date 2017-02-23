package ping

// ICMPPayload provides a somewhat secure signed payload that enables us to
// know when the package was sent without keeping local state.

import (
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"encoding/json"
	"errors"
	"time"

	"github.com/gansoi/gansoi/logger"
)

type (
	// ICMPPayload is capable of generating a signed ICMP payload.
	ICMPPayload struct {
		Helo      string    `json:"h"`
		Timestamp time.Time `json:"t"`
	}
)

var (
	secret = make([]byte, 16)
)

const (
	helo = "Gansoi ping agent"
)

func _init() {
	_, err := rand.Reader.Read(secret)
	if err != nil {
		logger.Info("ping", "Something is wrong with our random source: %s", err.Error())
	}
}

// NewICMPPayload returns a new ICMPPayload set to current time.
func NewICMPPayload() *ICMPPayload {
	return &ICMPPayload{
		Helo:      helo,
		Timestamp: time.Now(),
	}
}

// Bytes returns a signed payload.
func (p *ICMPPayload) Bytes() []byte {
	// THis should never fail.
	msg, _ := json.Marshal(p)

	h := md5.New()
	h.Write(secret)
	h.Write(msg)
	digest := h.Sum(nil)[:]

	return append(digest, msg...)
}

func (p *ICMPPayload) Read(payload []byte) error {
	if len(payload) < 16 {
		return errors.New("Payload too short")
	}

	h := md5.New()
	h.Write(secret)
	h.Write(payload[16:])
	digest := h.Sum(nil)[:]

	if !bytes.Equal(digest, payload[:16]) {
		return errors.New("Checksum error")
	}

	return json.Unmarshal(payload[16:], p)
}
