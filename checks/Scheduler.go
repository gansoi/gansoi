package checks

import (
	"time"

	"github.com/gansoi/gansoi/database"
	"github.com/gansoi/gansoi/logger"
	"github.com/gansoi/gansoi/stats"
)

type (
	// Scheduler takes care of scheduling checks on the local node. For now
	// it will spin four times each second.
	Scheduler struct {
		nodeName string
		stop     chan struct{}
		db       database.Database
	}
)

func init() {
	stats.CounterInit("scheduler_inflight")
	stats.CounterInit("scheduler_inflight_overrun")
	stats.CounterInit("scheduler_started")
	stats.CounterInit("scheduler_failed")
}

// NewScheduler instantiates a new scheduler.
func NewScheduler(db database.Database, nodeName string) *Scheduler {
	s := &Scheduler{
		nodeName: nodeName,
		stop:     make(chan struct{}),
		db:       db,
	}

	return s
}

// Run will start the event loop.
func (s *Scheduler) Run() {
	go s.loop()
}

// Stop will stop the event loop.
func (s *Scheduler) Stop() {
	s.stop <- struct{}{}
}

func (s *Scheduler) loop() {
	ticker := time.NewTicker(time.Millisecond * 250)

	for {
		select {
		case t := <-ticker.C:
			s.spin(t)
		case <-s.stop:
			ticker.Stop()
			return
		}
	}
}

func (s *Scheduler) spin(clock time.Time) {
	// Get a list of all checks.
	all, err := All(s.db)
	if err != nil {
		logger.Info("scheduler", "All() failed: %s", err.Error())
		return
	}

	for _, c := range all {
		meta := meta(clock, &c)

		// Calculate how much we should wait before executing the check. If
		// the value is positive, it's in the future.
		wait := meta.NextCheck.Sub(clock)

		// If there's still time to wait - wait.
		if wait > 0 {
			continue
		}

		// If the last check is still running, abort.
		if meta.running {
			stats.CounterInc("scheduler_inflight_overrun", 1)
			continue
		}

		meta.runs++
		meta.running = true

		stats.CounterInc("scheduler_inflight", 1)
		go s.runCheck(clock, c, meta)
	}
}

func (s *Scheduler) runCheck(clock time.Time, check Check, meta *checkMeta) *CheckResult {
	start := time.Now()

	stats.CounterInc("scheduler_started", 1)

	checkResult := RunCheck(&check)
	checkResult.Node = s.nodeName

	if checkResult.Error != "" {
		logger.Info("scheduler", "%s failed in %s: %s", check.ID, time.Since(start), checkResult.Error)
	} else {
		stats.CounterInc("scheduler_failed", 1)
		logger.Debug("scheduler", "%s ran in %s: %+v", check.ID, time.Since(start), checkResult.Results)
	}

	meta.NextCheck = clock.Add(check.Interval)

	s.db.Save(checkResult)

	meta.running = false

	return checkResult
}
