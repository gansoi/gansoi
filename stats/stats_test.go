package stats

import "testing"

func TestGetAll(t *testing.T) {
	CounterInit("TestGetAll")
	CounterSet("TestGetAll", 123)
	res := GetAll()

	v, ok := res["TestGetAll"].(int64)

	if !ok {
		t.Fatalf("GetAll() returned wrong type")
	}

	if v != 123 {
		t.Fatalf("GetAll() returned wrong value")
	}
}
