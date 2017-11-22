package plugins

import (
	"reflect"
	"strings"
)

type (
	// ArgumentDescription contains everything needed for creating a GUI for
	// configuring a plugin.
	ArgumentDescription struct {
		Name        string   `json:"name"`
		Type        string   `json:"type"`
		Description string   `json:"description"`
		Default     string   `json:"default"`
		EnumValues  []string `json:"enum"`
	}
)

func getArguments(elem reflect.Type) []ArgumentDescription {
	parameters := []ArgumentDescription{}

	l := elem.NumField()

	for i := 0; i < l; i++ {
		f := elem.Field(i)

		jsonName := f.Tag.Get("json")

		if f.Anonymous {
			parameters = append(parameters, getArguments(f.Type)...)
		} else if jsonName != "" {
			p := ArgumentDescription{}

			p.Name = jsonName
			p.Type = f.Type.String()
			p.Description = f.Tag.Get("description")
			p.Default = f.Tag.Get("default")
			enum := f.Tag.Get("enum")
			if enum != "" {
				p.EnumValues = strings.Split(enum, ",")
				p.Type = "enum"
			} else {
				p.EnumValues = []string{}
			}

			parameters = append(parameters, p)
		}
	}

	return parameters
}
