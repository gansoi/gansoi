package template

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"reflect"
	"strings"

	"honnef.co/go/js/dom"
)

var (
	// ErrTemplateNotFound will be returned if requested template could not be found.
	ErrTemplateNotFound = errors.New("Template not found")
)

type (
	// Collection is a colletion of templates.
	Collection struct {
		template *template.Template
	}
)

// NewCollection will instantiate a new collection of templates. It will read
// templates from script tags with the type templateType.
func NewCollection(templateType string) *Collection {
	root := template.New("root")

	c := &Collection{
		template: root,
	}

	doc := dom.GetWindow().Document()
	scripts := doc.GetElementsByTagName("script")
	for _, element := range scripts {
		typ := element.GetAttribute("type")

		if typ == templateType {
			name := element.Attributes()["name"]
			payload := element.InnerHTML()
			payload = strings.TrimSpace(payload)
			t := root.New(name)
			_, err := t.Parse(payload)

			if err != nil {
				fmt.Printf("%s\n", err.Error())
			}
		}
	}

	return c
}

// RenderString will render a template to a string.
func (c *Collection) RenderString(templateID string, data interface{}) (string, error) {
	t := c.template.Lookup(templateID)
	if t == nil {
		return "", fmt.Errorf("Template '%s' not found", templateID)
	}

	var buf bytes.Buffer
	err := t.Execute(&buf, data)

	return buf.String(), err
}

// formToMap will collect all input values and return a map.
func formToMap(form dom.Element) map[string]interface{} {
	values := make(map[string]interface{})

	inputs := form.GetElementsByTagName("input")
	for _, element := range inputs {
		// Type assertion should be safe here - but we check anyway.
		input, ok := element.(*dom.HTMLInputElement)
		if !ok {
			// This should never happen, since we're querying for "input".
			continue
		}

		if input.Name == "" {
			continue
		}

		switch input.Type {
		case "button":
			fallthrough
		case "submit":
			break

		case "number":
			values[input.Name] = input.ValueAsNumber

		case "text":
			values[input.Name] = input.Value

		case "checkbox":
			values[input.Name] = input.Checked

		default:
			panic("implement " + input.Type + " in formToMap()")
		}
	}

	return values
}

// RenderElement will render a template to a dom.Element.
func (c *Collection) RenderElement(target dom.Element, templateID string, data interface{}) error {
	r, ok := data.(Renderer)
	if ok {
		r.RenderFunc(func() error {
			return c.RenderElement(target, templateID, data)
		})
	}

	str, err := c.RenderString(templateID, data)
	if err != nil {
		return err
	}

	target.SetInnerHTML(str)

	immutable := reflect.ValueOf(data)

	// Find all buttons and attach methods
	buttons := target.GetElementsByTagName("button")
	for _, element := range buttons {
		// This should be safe
		button := element.(*dom.HTMLButtonElement)
		name := button.GetAttribute("name")

		// Find a method matching the buttons name
		val := immutable.MethodByName(name)
		if val.IsValid() {
			// If we have a valid match, attach the method to the button
			button.AddEventListener("click", true, func(event dom.Event) {
				// Get a value from the button to use as argument
				value := event.Target().GetAttribute("value")

				go val.Call([]reflect.Value{reflect.ValueOf(value)})
			})
		} else {
			return fmt.Errorf("No method for button %s", name)
		}
	}

	// Find all forms
	forms := target.GetElementsByTagName("form")
	for _, form := range forms {
		// Find all input elements
		inputs := form.GetElementsByTagName("input")
		for _, input := range inputs {
			// Get name attribute
			typ := input.GetAttribute("type")
			name := input.GetAttribute("name")

			val := immutable.Elem().FieldByName(name)

			switch typ {
			case "checkbox":
				if val.IsValid() && val.Kind() == reflect.Bool {
					input.(*dom.HTMLInputElement).Checked = val.Bool()
				}

			case "number":
				if val.IsValid() && val.Kind() == reflect.Int {
					input.(*dom.HTMLInputElement).Value = fmt.Sprintf("%d", val.Int())
				}

			case "email":
				fallthrough
			case "password":
				fallthrough
			case "url":
				fallthrough
			case "text":
				// text type should be pre-filled from data
				if val.IsValid() && val.Kind() == reflect.String {
					input.(*dom.HTMLInputElement).Value = val.String()
				}

			case "button":
				fallthrough
			case "submit":
				submitter, ok := data.(Submitter)
				if !ok {
					return fmt.Errorf("%T does not implement Submitter", data)
				}

				button := NewButton(input.(*dom.HTMLInputElement))

				input.AddEventListener("click", false, func(event dom.Event) {
					event.PreventDefault()

					// Disable button before calling handler. This makes sure
					// the user will not click twice.
					button.Disable()

					// Call Submit
					go submitter.Submit(button, formToMap(form))
				})

			default:
				return fmt.Errorf("Input type '%s' not supported", typ)
			}
		}
	}

	return err
}
