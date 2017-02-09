package plugins

import (
	"reflect"
)

type (
	// Notifier must be implemented by plugins capable of delivering
	// notifications.
	Notifier interface {
		Plugin

		Notify(text string) error
	}
)

var (
	notifiers = make(map[string]reflect.Type)
)

// RegisterNotifier will register the notifier with the notifier store.
func RegisterNotifier(name string, notifier interface{}) {
	_, found := notifiers[name]
	if found {
		// This should only happen at init time. panic() is okay for now.
		panic("A notifier with that name already exists")
	}

	notifiers[name] = reflect.TypeOf(notifier)
}

// GetNotifier will return a notifier registred with the name.
func GetNotifier(name string) Notifier {
	return reflect.New(notifiers[name]).Interface().(Notifier)
}

// ListNotifiers will return a list of all agents.
func ListNotifiers() []PluginDescription {
	list := make([]PluginDescription, 0, len(agents))

	for name, typ := range notifiers {
		list = append(list, PluginDescription{
			Name:      name,
			Arguments: getArguments(typ),
		})
	}

	return list
}
