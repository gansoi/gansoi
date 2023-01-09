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

	if e, ok := data.(*eval.Evaluation); ok {
		err := n.gotEvaluation(e)
		if err != nil {
			logger.Debug("notify", "%s", err.Error())
		}
	}
}

func (n *Notifier) gotEvaluation(e *eval.Evaluation) error {
	var check checks.Check

	err := n.db.One("ID", e.CheckID, &check)
	if err != nil {
		return err
	}

	// For how long have we had this state?
	duration := e.End.Sub(e.Start)

	// Retrieve the last known state of the check. If the last state is
	// unknown, StateUnknown will be used.
	stateCacheLock.RLock()
	lastState := stateCache[e.CheckHostID]
	stateCacheLock.RUnlock()

	// If nothing changed since last evaluation, we can safely abort since
	// there's nothing to notify about.
	if e.State == lastState {
		logger.Debug("notify", "[%s] Ignoring unchanged state (Last: %s, Current: %s, Duration: %s) %s", e.CheckHostID, lastState, e.State, duration.String(), e.History.ColorString())
		return nil
	}

	stateCacheLock.Lock()
	stateCache[e.CheckHostID] = e.State
	stateCacheLock.Unlock()

	// If we arrive here we know that state has changed since last evaluation.
	// If we changed from StateUnknown, we ignore this state change because it
	// is caused by a check "coming online".
	if lastState == eval.StateUnknown {
		logger.Info("notify", "[%s] Ignoring %s when previous state is %s %v", e.CheckHostID, e.State, lastState, e.History.ColorString())

		return nil
	}

	logger.Info("notify", "%s is %s %s", e.CheckHostID, e.State, e.History.ColorString())

	text := fmt.Sprintf("%s is %s", e.CheckHostID, e.State.String())

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

			contact.Notify(text)
		}
	}

	return nil
}
