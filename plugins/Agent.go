package plugins

import "reflect"

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

// Get will return an agent registered with the name.
func GetAgent(name string) interface{} {
	agent, found := agents[name]
	if !found {
		return nil
	}

	return reflect.New(agent).Interface()
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
