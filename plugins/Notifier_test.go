package plugins

import "testing"

type (
	mockNotifier struct {
		Err error `json:"err" description:"The error to return from Notify()"`
	}
)

func init() {
	RegisterNotifier("mockNotifier", mockNotifier{})
}

func (n *mockNotifier) Notify(text string) error {
	return n.Err
}

func TestNotifierRegisterDouble(t *testing.T) {
	defer mustPanic(t)

	RegisterNotifier("mockagentdouble", mockNotifier{})
	RegisterNotifier("mockagentdouble", mockNotifier{})
}

func TestNotifierGet(t *testing.T) {
	const id = "mockagent_TestNotifierGet"

	RegisterNotifier(id, mockNotifier{})

	agent := GetNotifier(id)
	if agent == nil {
		t.Fatalf("GetNotifier() returned nil")
	}
}

func TestNotifierGetFail(t *testing.T) {
	agent := GetNotifier("nonexistingagent_TestNotifierGetFail")
	if agent != nil {
		t.Fatalf("GetNotifier() did not return nil")
	}
}

func TestListNotifiers(t *testing.T) {
	found := false
	list := ListNotifiers()

	for _, d := range list {
		if d.Name == "mockNotifier" {
			found = true
		}
	}

	if !found {
		t.Fatalf("mockNotifier not found in list")
	}
}
