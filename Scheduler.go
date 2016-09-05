package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/abrander/gansoi/node"
)

type (
	// Scheduler takes care of scheduling checks.
	Scheduler struct {
		run       bool
		node      *node.Node
		ticker    *time.Ticker
		metaStore map[string]*checkMeta
	}

	checkMeta struct {
		LastCheck time.Time
		NextCheck time.Time
	}
)

// NewScheduler starts a new scheduler.
func NewScheduler(n *node.Node, run bool) *Scheduler {
	s := &Scheduler{
		node:      n,
		ticker:    time.NewTicker(time.Millisecond * 1000),
		run:       run,
		metaStore: make(map[string]*checkMeta),
	}

	go s.loop()

	return s
}

// Run will start the event loop.
func (s *Scheduler) Run() {
	s.run = true
}

// Stop will stp the event loop.
func (s *Scheduler) Stop() {
	s.run = false
}

func (s *Scheduler) meta(check *Check) *checkMeta {
	meta, found := s.metaStore[check.ID]
	if !found {
		meta = &checkMeta{}
		s.metaStore[check.ID] = meta
	}

	return meta
}

func (s *Scheduler) loop() {
	// inFlight is a list of check id's currently running
	inFlight := make(map[string]bool)
	inFlightLock := sync.RWMutex{}

	for t := range s.ticker.C {
		if !s.run {
			continue
		}

		// We start by extracting a list of all checks. If this gets too
		// expensive at some point, we can do it less frequent or more
		// efficient.
		var allChecks []Check
		err := s.node.All(&allChecks, -1, 0, false)
		if err != nil {
			fmt.Printf("Err: %s\n", err.Error())
			continue
		}

		// We iterate the list of checks, to see if anything needs to be done.
		for _, check := range allChecks {
			meta := s.meta(&check)

			// Calculate the age of the last check, if the age is positive, it's
			// in the past.
			age := t.Sub(meta.LastCheck)

			// Calculate how much we should wait before executing the check. If
			// the value is positive, it's in the future.
			wait := meta.NextCheck.Sub(t)

			// Check if the check is already executing.
			inFlightLock.RLock()
			_, found := inFlight[check.ID]
			inFlightLock.RUnlock()

			if found {
				continue
			}

			// If the check is older than two intervals, we treat it as new.
			if age > check.Interval*2 && wait < -check.Interval {
				checkIn := time.Duration(rand.Int63n(int64(check.Interval)))
				meta.NextCheck = t.Add(checkIn)

				fmt.Printf("%s start delayed for %s\n", check.ID, checkIn.String())
			} else if wait < 0 {
				// If we arrive here, wait is sub-zero, which means that we
				// should execute now.
				inFlightLock.Lock()
				inFlight[check.ID] = true
				inFlightLock.Unlock()

				// Execute the check in its own go routine.
				go func(check Check) {
					// Run the job.
					start := time.Now()

					result, err := check.Agent.Check()

					if err != nil {
						fmt.Printf("%s failed in %s: %s\n", check.ID, time.Now().Sub(start), err.Error())
					} else {
						fmt.Printf("%s ran in %s: %+v\n", check.ID, time.Now().Sub(start), result)

						s.node.SubmitResult(check.ID, err, result)
					}

					// Save the check time and schedule next check.
					meta.LastCheck = t
					meta.NextCheck = t.Add(check.Interval)

					// Remove the check from the inFlight map.
					inFlightLock.Lock()
					delete(inFlight, check.ID)
					inFlightLock.Unlock()
				}(check)
			}
		}
	}
}
