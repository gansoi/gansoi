package plugins

import (
	"reflect"
)

type (
	// Agent should be implemented by all agents. An agent is the entity
	// responsible for carrying out all checks.
	Agent interface {
		Plugin

		// Check should run the agents check.
		Check(result AgentResult) error
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

// GetAgent will return an agent registred with the name.
func GetAgent(name string) Agent {
	return reflect.New(agents[name]).Interface().(Agent)
}

// ListAgents will return a list of all agents.
func ListAgents() []PluginDescription {
	list := make([]PluginDescription, 0, len(agents))

	for name, typ := range agents {
		list = append(list, PluginDescription{
			Name:      name,
			Arguments: getArguments(typ),
		})
	}

	return list
}
