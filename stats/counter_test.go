package stats

import (
	"runtime"
	"sync"
	"testing"
)

func mustPanic(t *testing.T) {
	if r := recover(); r == nil {
		t.Errorf("Unexisting counter did not cause a panic")
	}
}

func TestCounterInit(t *testing.T) {
	CounterInit("initialized")
	CounterGet("initialized")
}

func TestCounterInitPanic(t *testing.T) {
	defer mustPanic(t)

	CounterInit("initialized")
	CounterInit("initialized")
}

func TestCounterInc(t *testing.T) {
	CounterInit("TestCounterInc")
	CounterInc("TestCounterInc", 1)
	CounterInc("TestCounterInc", 1)
	CounterInc("TestCounterInc", 1)
	if CounterGet("TestCounterInc") != 3 {
		t.Fatalf("CounterInc failed")
	}
}

func TestCounterIncPanic(t *testing.T) {
	defer mustPanic(t)

	CounterInc("not initialized", 1)
}

func TestCounterSetGet(t *testing.T) {
	CounterInit("TestCounterGet")
	CounterSet("TestCounterGet", 123)
	if CounterGet("TestCounterGet") != 123 {
		t.Fatalf("CounterSet or CounterGet failed")
	}
}

func TestCounterGetPanic(t *testing.T) {
	defer mustPanic(t)

	CounterGet("not initialized")
}

func TestCounterSetPanic(t *testing.T) {
	defer mustPanic(t)

	CounterSet("not initialized", 0)
}

func TestCounterConcurrency(t *testing.T) {
	const con = 1000

	var sum int64
	for i := int64(0); i < 500; i++ {
		sum += i
	}

	prev := runtime.GOMAXPROCS(20)

	CounterInit("TestCounterConcurrency")

	var wg sync.WaitGroup
	wg.Add(con)

	for n := 0; n < con; n++ {
		go func() {
			for i := 0; i < 500; i++ {
				CounterInc("TestCounterConcurrency", int64(i))
			}
			wg.Done()
		}()
	}
	wg.Wait()

	if CounterGet("TestCounterConcurrency") != sum*con {
		t.Fatalf("Concurrencytest failed, expected %d, got %d", sum*con, CounterGet("TestCounterConcurrency"))
	}

	runtime.GOMAXPROCS(prev)
}
