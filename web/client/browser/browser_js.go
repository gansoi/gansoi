// +build js

package browser

import (
	"honnef.co/go/js/dom"
)

func init() {
	win = dom.GetWindow()
	doc = win.Document()
}
