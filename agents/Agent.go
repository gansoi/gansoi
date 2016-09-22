package agents

import (
	"reflect"
	"strings"
)

type (
	// Agent should be implemented by all agents. An agent is the entity
	// responsible for carrying out all checks.
	Agent interface {
		// Check should run the agents check.
		Check() (interface{}, error)
	}

	// ArgumentDescription contains everything needed for creating a GUI for
	// configuring an agent.
	ArgumentDescription struct {
		Name        string   `json:"name"`
		Type        string   `json:"type"`
		Description string   `json:"description"`
		EnumValues  []string `json:"enum"`
	}

	// AgentDescription describes an agent.
	AgentDescription struct {
		Name      string                `json:"name"`
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

// GetAgent will return an agent registred with the name.
func GetAgent(name string) Agent {
	return reflect.New(agents[name]).Interface().(Agent)
}

// ListAgents will return a list of all agents.
func ListAgents() []AgentDescription {
	list := make([]AgentDescription, 0, len(agents))

	for name, typ := range agents {
		list = append(list, AgentDescription{
			Name:      name,
			Arguments: getArguments(typ),
		})
	}

	return list
}

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
