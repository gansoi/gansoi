package browser

import (
	"github.com/gopherjs/gopherjs/js"
	"honnef.co/go/js/dom"
)

type (
	HashChangeListener struct {
		C        chan string
		listener func(*js.Object)
	}
)

func NewHashChangeListener() *HashChangeListener {
	h := &HashChangeListener{
		C: make(chan string),
	}

	h.listener = win.AddEventListener("hashchange", true, h.change)

	return h
}

// When the listener is done watching, this should be called to remove
// the event listener
func (h *HashChangeListener) Close() {
	win.RemoveEventListener("hashchange", true, h.listener)
}

// Should be called from the "hashchange" callback
func (h *HashChangeListener) change(event dom.Event) {
	// We ignore the values present in the dom.HashChangeEvent and
	// simply call Url()
	h.C <- Url().Fragment
}
