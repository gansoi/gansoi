package plugins

import (
	"reflect"
)

type (
	// Notifier must be implemented by plugins capable of delivering
	// notifications.
	Notifier interface {
		Notify(text string) error
	}

	// NotifierDescription describes a notifier.
	NotifierDescription struct {
		Name      string                `json:"name"`
		Arguments []ArgumentDescription `json:"arguments"`
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
	notifier, found := notifiers[name]
	if !found {
		return nil
	}

	return reflect.New(notifier).Interface().(Notifier)
}

// ListNotifiers will return a list of all agents.
func ListNotifiers() []NotifierDescription {
	list := make([]NotifierDescription, 0, len(agents))

	for name, typ := range notifiers {
		list = append(list, NotifierDescription{
			Name:      name,
			Arguments: getArguments(typ),
		})
	}

	return list
}
