package ca

import (
	"crypto/rand"
	"errors"
)

type (
	failReader struct {
		called int
		failAt int
	}
)

func (r *failReader) Read(p []byte) (n int, err error) {
	r.called++

	if r.called == r.failAt {
		return 0, errors.New("failReader fails")
	}

	return rand.Read(p)
}
