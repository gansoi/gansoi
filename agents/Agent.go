package agents

import "reflect"

type (
	// Agent should be implemented by all agents.
	Agent interface {
		// Check should run the agents check.
		Check() (interface{}, error)
	}
)

var (
	agents = make(map[string]reflect.Type)
)

// RegisterAgent will register the agent with the agent store.
func RegisterAgent(name string, agent interface{}) {
	_, found := agents[name]
	if found {
		panic("An agent with that name already exists")
	}

	agents[name] = reflect.TypeOf(agent)
}

// GetAgent will return an agent registred with the name.
func GetAgent(name string) Agent {
	return reflect.New(agents[name]).Interface().(Agent)
}
