package checks

import (
	"time"

	"github.com/gansoi/gansoi/database"
	"github.com/gansoi/gansoi/logger"
	"github.com/gansoi/gansoi/stats"
	"github.com/gansoi/gansoi/transports"
	"github.com/gansoi/gansoi/transports/ssh"
)

type (
	// Scheduler takes care of scheduling checks on the local node. For now
	// it will spin four times each second.
	Scheduler struct {
		nodeName string
		stop     chan struct{}
		db       database.Database
		store    *MetaStore
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
	store, _ := newMetaStore(db)

	s := &Scheduler{
		nodeName: nodeName,
		stop:     make(chan struct{}),
		db:       db,
		store:    store,
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
	meta := s.store.Next(clock)
	if meta == nil {
		return
	}

	stats.CounterInc("scheduler_inflight", 1)
	go s.runCheck(clock, meta)
}

func (s *Scheduler) runCheck(clock time.Time, meta *checkMeta) *CheckResult {
	start := time.Now()

	stats.CounterInc("scheduler_started", 1)

	var checkResult *CheckResult
	var transport transports.Transport

	if meta.key.hostID != "" {
		remote := ssh.SSH{}
		err := s.db.One("ID", meta.key.hostID, &remote)
		if err == nil {
			transport = &remote
		}
	}

	checkResult = RunCheck(transport, meta.check)
	checkResult.HostID = meta.key.hostID

	stats.CounterInc("scheduler_inflight", -1)

	checkResult.Node = s.nodeName

	if checkResult.Error != "" {
		stats.CounterInc("scheduler_failed", 1)
		logger.Info("scheduler", "%s failed in %s: %s", meta.check.ID, time.Since(start), checkResult.Error)
	} else {
		logger.Debug("scheduler", "%s ran in %s: %+v", meta.check.ID, time.Since(start), checkResult.Results)
	}

	s.db.Save(checkResult)

	s.store.Done(meta)

	return checkResult
}
