package browser

import (
	"net/url"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"
	"honnef.co/go/js/dom"
)

type (
	Renderer interface {
		RenderElement(target dom.Element, id string, data interface{}) error
	}
)

var (
	win dom.Window
	doc dom.Document
)

func ID(id string) dom.Element {
	return doc.GetElementByID(id)
}

func Url() *url.URL {
	location := js.Global.Get("location")

	// This should never fail. Well.
	u, _ := url.Parse(location.Get("href").String())

	return u
}

func SetUrl(u url.URL) {
	location := js.Global.Get("location")

	location.Set("href", u.String())
}

func WaitForLoad() {
	d := js.Global.Get("document")
	readyState := d.Get("readyState").String()

	// If we're already loaded, simply exit
	if readyState == "complete" {
		return
	}

	// hackery to wait for documentready
	ch := make(chan bool)
	win.AddEventListener("load", true, func(event dom.Event) {
		ch <- true
	})
	<-ch
}

func ShowModal(renderer Renderer, id string, data interface{}) error {
	html := `<div id="myModal" class="modal fade" tabindex="-1" role="dialog">
  <div class="modal-dialog">
    <div class="modal-content">
      <div class="modal-header">
        <button type="button" class="close" data-dismiss="modal" aria-label="Close"><span aria-hidden="true">&times;</span></button>
        <h4 class="modal-title">Modal title</h4>
      </div>
      <div class="modal-body">
      </div>
      <div class="modal-footer">
        <button type="button" class="btn btn-default" data-dismiss="modal">Close</button>
        <button type="button" class="btn btn-primary">Save changes</button>
      </div>
    </div>
  </div>
</div>`

	div := doc.CreateElement("div")
	div.SetInnerHTML(html)

	body := doc.GetElementsByTagName("body")
	body[0].AppendChild(div.FirstChild())

	targets := div.GetElementsByClassName("modal-body")
	err := renderer.RenderElement(targets[0], id, data)
	if err != nil {
		return err
	}

	// Bootstrap attaches its methods to jQuery, so we have to use that
	m := jquery.NewJQuery(div.FirstChild())
	m.Call("modal")

	return nil
}
