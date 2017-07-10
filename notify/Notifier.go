package notify

import (
	"expvar"
	"fmt"
	"sync"

	"github.com/gansoi/gansoi/checks"
	"github.com/gansoi/gansoi/database"
	"github.com/gansoi/gansoi/eval"
	"github.com/gansoi/gansoi/logger"
)

type (
	// Notifier takes care of notifying all contacts in a relevant ContactGroup.
	Notifier struct {
		db database.Reader
	}
)

var (
	stateCacheLock sync.RWMutex
	stateCache     = make(map[string]eval.State)

	sent = expvar.NewInt("notification_sent")
)

// NewNotifier will start a new notifier service.
func NewNotifier(db database.Reader) (*Notifier, error) {
	n := &Notifier{
		db: db,
	}

	return n, nil
}

// PostApply implements database.Listener.
func (n *Notifier) PostApply(leader bool, command database.Command, data interface{}) {
	if !leader {
		return
	}

	switch data.(type) {
	case *eval.Evaluation:
		e := n.gotEvaluation(data.(*eval.Evaluation))
		if e != nil {
			logger.Debug("notify", "%s", e.Error())
		}
	}
}

func (n *Notifier) gotEvaluation(e *eval.Evaluation) error {
	var check checks.Check
	err := n.db.One("ID", e.CheckHostID, &check)
	if err != nil {
		return err
	}

	state := e.History.Last(3).Reduce()

	if state == eval.StateUnknown {
		logger.Debug("notify", "[%s] Ignoring %s %s", e.CheckHostID, state.String(), e.History.ColorString())
		return nil
	}

	duration := e.End.Sub(e.Start)

	stateCacheLock.RLock()
	lastState := stateCache[e.CheckHostID]
	stateCacheLock.RUnlock()

	if state == lastState {
		logger.Debug("notify", "[%s] Ignoring unchanged state (Last: %s, Current: %s, Duration: %s) %s", e.CheckHostID, lastState, state, duration.String(), e.History.ColorString())
		// Nothing changed. Abort.
		return nil
	}

	stateCacheLock.Lock()
	stateCache[e.CheckHostID] = state
	stateCacheLock.Unlock()

	if lastState == eval.StateUnknown {
		logger.Info("notify", "[%s] Ignoring %s when previous state is %s %v", e.CheckHostID, state, lastState, e.History.ColorString())
		// Last state was unknown. Maybe we just started. Don't notify.
		return nil
	}

	if duration < check.Interval*2 && state == eval.StateDegraded {
		logger.Info("notify", "[%s] Ignoring %s for less than two cycles when degraded %s", e.CheckHostID, state, e.History.ColorString())
		return nil
	}

	logger.Info("notify", "%s is %s %s", e.CheckHostID, state, e.History.ColorString())

	if len(check.ContactGroups) == 0 {
		logger.Info("notify", "[%s] No-one to nofity, aborting", e.CheckHostID)
		return nil
	}

	text := fmt.Sprintf("%s is %s", e.CheckHostID, state.String())

	targetGroups := check.ContactGroups
	for _, groupID := range targetGroups {
		group, err := LoadContactGroup(n.db, groupID)
		if err != nil {
			logger.Info("notify", "[%s] ContactGroup not found (%s)", e.CheckHostID, groupID)
			continue
		}

		contacts, _ := group.GetContacts(n.db)
		for _, contact := range contacts {
			sent.Add(1)
			logger.Info("notify", "[%s] Notifying '%s' using %s", e.CheckHostID, contact.ID, contact.Notifier)

			go contact.Notify(text)
		}
	}

	return nil
}
