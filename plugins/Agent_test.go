package plugins

import "testing"

type (
	mockAgent struct {
		Err               error  `json:"err" description:"The error to return from Check()"`
		Something         string `default:"testing"`
		SomethingInt      int    `default:"-23"`
		SomethingUint     uint   `default:"48"`
		SomethingBooleanA bool   `default:"true"`
		SomethingBooleanB bool   `default:"false"`
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

	m := agent.(*mockAgent)
	if m.Something != "testing" {
		t.Errorf("GetAgent() failed to set default Something, got '%s'", m.Something)
	}

	if m.SomethingInt != -23 {
		t.Errorf("GetAgent() failed to set default for SomethingInt, got %d", m.SomethingInt)
	}

	if m.SomethingUint != 48 {
		t.Errorf("GetAgent() failed to set default for SomethingUint, got %d", m.SomethingUint)
	}

	if m.SomethingBooleanA != true {
		t.Errorf("GetAgent() failed to set default for SomethingBooleanA, got %v", m.SomethingBooleanA)
	}

	if m.SomethingBooleanB != false {
		t.Errorf("GetAgent() failed to set default for SomethingBooleanB, got %v", m.SomethingBooleanB)
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
