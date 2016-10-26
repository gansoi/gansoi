package template

import (
	"honnef.co/go/js/dom"
)

type (
	// Button is a helper type for Submitter's.
	Button struct {
		element *dom.HTMLInputElement
	}
)

// NewButton will instantiate a new Button from button.
func NewButton(button *dom.HTMLInputElement) *Button {
	return &Button{element: button}
}

// Enable will enable the button in the UI.
func (b *Button) Enable() {
	b.element.Disabled = false
}

// Disable will render the button disabled in the dom.
func (b *Button) Disable() {
	b.element.Disabled = true
}

// Name returns the name gived to a button in the name property.
func (b *Button) Name() string {
	return b.element.Name
}

// Index returns the content of the "index" attribute. Will return "" if not
// set.
func (b *Button) Index() string {
	return b.element.Attributes()["index"]
}
