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

func formCollect(form dom.Element, v reflect.Value) {
	mutable := v.Elem()

	inputs := form.GetElementsByTagName("input")
	for _, element := range inputs {
		// Type assertion should be safe here
		input := element.(*dom.HTMLInputElement)

		// Get name attribute
		name := input.GetAttribute("name")
		val := mutable.FieldByName(name)

		// See if we found a matching field
		if val.IsValid() {
			// We only support string for now
			if val.Kind() != reflect.String {
				return
			}

			val.Set(reflect.ValueOf(input.Value))
		}
	}
}

// RenderElement will render a template to a dom.Element.
func (c *Collection) RenderElement(target dom.Element, templateID string, data interface{}) error {
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

			switch typ {
			case "email":
				fallthrough
			case "password":
				fallthrough
			case "url":
				fallthrough
			case "text":
				// text type should be pre-filled from data
				val := immutable.Elem().FieldByName(name)
				if !val.IsValid() {
					return fmt.Errorf("Did not find field with name '%s' for '%s' input", name, typ)
				}

				// We only support string for now
				if val.Kind() != reflect.String {
					return errors.New("Only string supported for now")
				}

				input.(*dom.HTMLInputElement).Value = val.String()
			case "submit":
				fun := immutable.MethodByName(name)
				if !fun.IsValid() {
					return fmt.Errorf("Did not find submit method with name '%s'", name)
				}

				// On submit we should collect all values
				input.AddEventListener("click", false, func(event dom.Event) {
					event.PreventDefault()

					formCollect(form, immutable)

					go fun.Call([]reflect.Value{})
				})
			default:
				return fmt.Errorf("Input type '%s' not supported", typ)
			}
		}
	}

	return err
}
