package plugins

import (
	"reflect"
	"strconv"
)

type (
	// Agent should be implemented by all agents. An agent is the entity
	// responsible for carrying out all checks.
	Agent interface {
		// Check should run the agents check.
		Check(result AgentResult) error
	}

	// AgentDescription describes an agent.
	AgentDescription struct {
		Name      string                `json:"name"`
		Remote    bool                  `json:"remote"`
		Arguments []ArgumentDescription `json:"arguments"`
	}
)

var (
	agents = make(map[string]reflect.Type)
)

// RegisterAgent will register the agent with the agent store.
func RegisterAgent(name string, agent interface{}) {
	_, found := agents[name]
	if found {
		// This should only happen at init time. panic() is okay for now.
		panic("An agent with that name already exists")
	}

	agents[name] = reflect.TypeOf(agent)
}

// GetAgent will return an agent registered with the name.
func GetAgent(name string) interface{} {
	agent, found := agents[name]
	if !found {
		return nil
	}

	a := reflect.New(agent)

	setDefault(agent, a)

	return a.Interface()
}

// ListAgents will return a list of all agents.
func ListAgents() []AgentDescription {
	list := make([]AgentDescription, 0, len(agents))

	for name, typ := range agents {
		_, remote := reflect.New(typ).Interface().(RemoteAgent)

		list = append(list, AgentDescription{
			Name:      name,
			Remote:    remote,
			Arguments: getArguments(typ),
		})
	}

	return list
}

// setString will try to set a reflect.Value based on a string representation
// of the value.
func setString(v reflect.Value, str string) {
	switch v.Type().Kind() {
	case reflect.String:
		v.SetString(str)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		value, _ := strconv.ParseInt(str, 10, 64)
		v.SetInt(value)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		value, _ := strconv.ParseUint(str, 10, 64)
		v.SetUint(value)
	}
}

// setDefault will set default values for a reflect value based on the struct
// tag "default".
func setDefault(t reflect.Type, v reflect.Value) {
	l := t.NumField()
	v = v.Elem()

	for i := 0; i < l; i++ {
		f := t.Field(i)

		def := f.Tag.Get("default")
		if def == "" {
			continue
		}

		setString(v.Field(i), def)
	}
}
