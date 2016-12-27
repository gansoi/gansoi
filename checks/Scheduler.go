package checks

import (
	"math/rand"
	"sync"
	"time"

	"github.com/gansoi/gansoi/database"
	"github.com/gansoi/gansoi/logger"
	"github.com/gansoi/gansoi/stats"
)

type (
	// Scheduler takes care of scheduling checks on the local node. For now
	// it will tick two times each second.
	Scheduler struct {
		run       bool
		node      database.Database
		nodeName  string
		ticker    *time.Ticker
		metaStore map[string]*checkMeta
	}

	// checkMeta is used internally in the scheduler to keep track of check
	// metadata.
	checkMeta struct {
		LastCheck time.Time
		NextCheck time.Time
	}
)

func init() {
	stats.CounterInit("scheduler_inflight")
	stats.CounterInit("scheduler_inflight_overrun")
	stats.CounterInit("scheduler_started")
	stats.CounterInit("scheduler_failed")
}

// NewScheduler starts a new scheduler.
func NewScheduler(n database.Database, nodeName string, run bool) *Scheduler {
	s := &Scheduler{
		node:      n,
		nodeName:  nodeName,
		ticker:    time.NewTicker(time.Millisecond * 500),
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
			logger.Red("scheduler", "%s", err.Error())
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
				stats.CounterInc("scheduler_inflight_overrun", 1)
				continue
			}

			// If the check is older than two intervals, we treat it as new.
			if age > check.Interval*2 && wait < -check.Interval {
				checkIn := time.Duration(rand.Int63n(int64(check.Interval)))
				meta.NextCheck = t.Add(checkIn)

				logger.Green("scheduler", "%s start delayed for %s", check.ID, checkIn.String())
			} else if wait < 0 {
				// If we arrive here, wait is sub-zero, which means that we
				// should execute now.
				inFlightLock.Lock()
				inFlight[check.ID] = true
				inFlightLock.Unlock()
				stats.CounterInc("scheduler_inflight", 1)

				// Execute the check in its own go routine.
				go func(check Check) {
					// Run the job.
					start := time.Now()

					stats.CounterInc("scheduler_started", 1)
					checkResult := RunCheck(&check)
					checkResult.Node = s.nodeName

					if checkResult.Error != "" {
						logger.Yellow("scheduler", "%s failed in %s: %s", check.ID, time.Now().Sub(start), checkResult.Error)
					} else {
						stats.CounterInc("scheduler_failed", 1)
						logger.Green("scheduler", "%s ran in %s: %+v", check.ID, time.Now().Sub(start), checkResult.Results)
					}

					s.node.Save(checkResult)

					// Save the check time and schedule next check. It should be
					// safe to update meta from a go routine. We shouldn't
					// execute the same check more than once at a time.
					meta.LastCheck = t
					meta.NextCheck = t.Add(check.Interval)

					// Remove the check from the inFlight map.
					inFlightLock.Lock()
					delete(inFlight, check.ID)
					inFlightLock.Unlock()

					stats.CounterDec("scheduler_inflight", 1)
				}(check)
			}
		}
	}
}
