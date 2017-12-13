package checks

import (
	"math/rand"
	"sync"
	"time"

	"github.com/gansoi/gansoi/database"
)

type (
	// MetaStore will keep a list of checks to execute.
	MetaStore struct {
		sync.RWMutex
		store map[metaKey]*checkMeta
	}

	metaKey struct {
		checkID string
		hostID  string
	}

	checkMeta struct {
		check     Check
		key       *metaKey
		running   bool
		runs      int
		NextCheck time.Time
		interval  time.Duration
	}
)

func newMetaStore(db database.ReadWriteBroadcaster) (*MetaStore, error) {
	s := &MetaStore{
		store: make(map[metaKey]*checkMeta),
	}

	db.RegisterListener(s)

	s.populate(db)

	return s, nil
}

// PostApply implements database.Listener.
func (s *MetaStore) PostApply(leader bool, command database.Command, data interface{}) {
	clock := time.Now()

	check, isCheck := data.(*Check)
	if !isCheck {
		return
	}

	// We have no way of knowing if the check is new or not, so we have to
	// always delete the known metadata.
	s.removeCheck(check)

	if command == database.CommandSave {
		s.addCheck(clock, check)
	}
}

func (s *MetaStore) populate(db database.Reader) error {
	clock := time.Now()
	var allChecks []Check
	err := db.All(&allChecks, -1, 0, false)
	if err != nil {
		return err
	}

	for _, c := range allChecks {
		s.addCheck(clock, &c)
	}

	return nil
}

func (s *MetaStore) removeCheck(check *Check) {
	s.Lock()

	for key := range s.store {
		if key.checkID == check.ID {
			delete(s.store, key)
		}
	}

	s.Unlock()
}

func (s *MetaStore) addCheck(clock time.Time, check *Check) {
	if len(check.Hosts) == 0 {
		key := metaKey{
			checkID: check.ID,
		}

		meta := &checkMeta{
			NextCheck: randomStartTime(clock, check.Interval),
			interval:  check.Interval,
			check:     *check,
			key:       &key,
		}

		s.Lock()
		s.store[key] = meta
		s.Unlock()
	}

	for _, hostID := range check.Hosts {
		key := metaKey{
			checkID: check.ID,
			hostID:  hostID,
		}

		meta := &checkMeta{
			NextCheck: randomStartTime(clock, check.Interval),
			interval:  check.Interval,
			check:     *check,
			key:       &key,
		}

		s.Lock()
		s.store[key] = meta
		s.Unlock()
	}
}

// next returns the next check to execute. Done() must be called when the check
// is done executing.
func (s *MetaStore) next(clock time.Time) *checkMeta {
	var winner *checkMeta

	s.RLock()
	for _, meta := range s.store {
		// Calculate how much we should wait before executing the check. If
		// the value is positive, it's in the future.
		wait := meta.NextCheck.Sub(clock)

		// ... and if the value is negative, we better get on with it :)
		if wait < 0 {
			if meta.running {
				inflightOverrun.Add(1)
				continue
			}

			winner = meta

			break
		}
	}

	s.RUnlock()

	if winner != nil {
		// We lock the MetaStore just to change a single checkMeta. That is a
		// bit excessive, but it'll do for now.
		s.Lock()
		winner.runs++
		winner.running = true
		winner.NextCheck = clock.Add(winner.interval)
		s.Unlock()
	}

	return winner
}

// Done must be called to signal that a check is done. After Done() the
// checkMeta can again be returned from Next().
func (s *MetaStore) Done(meta *checkMeta) {
	s.Lock()
	meta.running = false
	s.Unlock()
}

// randomStartTime will try to find a random start time for a check. It should
// be randomized to distribute load for heavy checks. We could be checking 100
// frontend servers that will all hit the same backend database - and we want
// to monitor downtime, not create downtime by DOS'ing :)
func randomStartTime(clock time.Time, delay time.Duration) time.Time {
	// If the delay is more than one minute, we clamp it. First check should
	// be fairly quick.
	if delay > time.Second*60 {
		delay = time.Second * 60
	}

	// This will not distribute all checks evenly in the "delay" space, but
	// it's good enough for now.
	delay = time.Duration(rand.Int63n(int64(delay)))

	return clock.Add(delay)
}
