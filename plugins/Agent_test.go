package plugins

import "testing"

type (
	mockAgent struct {
		Err error `json:"err" description:"The error to return from Check()"`
	}
)

func init() {
	RegisterAgent("mockAgent", mockAgent{})
}

func (a *mockAgent) Check(result AgentResult) error {
	return a.Err
}

func TestAgentRegisterDouble(t *testing.T) {
	defer mustPanic(t)

	RegisterAgent("mockagentdouble", mockAgent{})
	RegisterAgent("mockagentdouble", mockAgent{})
}

func TestAgentGet(t *testing.T) {
	const id = "mockagent_TestAgentGet"

	RegisterAgent(id, mockAgent{})

	agent := GetAgent(id).(Agent)
	if agent == nil {
		t.Fatalf("GetAgent() returned nil")
	}
}

func TestAgentGetFail(t *testing.T) {
	agent := GetAgent("nonexistingagent_TestAgentGetFail")
	if agent != nil {
		t.Fatalf("GetAgent() did not return nil")
	}
}

func TestListAgents(t *testing.T) {
	found := false
	list := ListAgents()

	for _, d := range list {
		if d.Name == "mockAgent" {
			found = true
		}
	}

	if !found {
		t.Fatalf("mockAgent not found in list")
	}
}
