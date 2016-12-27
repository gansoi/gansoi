package notify

import (
	"fmt"
	"sync"

	"github.com/gansoi/gansoi/checks"
	"github.com/gansoi/gansoi/database"
	"github.com/gansoi/gansoi/eval"
	"github.com/gansoi/gansoi/logger"
	"github.com/gansoi/gansoi/stats"
)

type (
	// Notifier takes care of notifying all contacts in a relevant ContactGroup.
	Notifier struct {
		db database.Database
	}
)

var (
	cacheLock sync.RWMutex

	contactsCache = make(map[string]*Contact)
	groupsCache   = make(map[string]*ContactGroup)

	stateCacheLock sync.RWMutex
	stateCache     = make(map[string]eval.State)
)

func init() {
	stats.CounterInit("notification_sent")
}

// NewNotifier will start a new notifier service.
func NewNotifier(db database.Database) (*Notifier, error) {
	n := &Notifier{
		db: db,
	}

	var contacts []Contact
	var groups []ContactGroup

	err := n.db.All(&contacts, -1, 0, false)
	if err != database.ErrNotFound && err != nil {
		return nil, err
	}

	err = n.db.All(&groups, -1, 0, false)
	if err != database.ErrNotFound && err != nil {
		return nil, err
	}

	cacheLock.Lock()
	for _, contact := range contacts {
		contactsCache[contact.ID] = &contact
	}

	for _, group := range groups {
		groupsCache[group.ID] = &group
	}
	cacheLock.Unlock()

	return n, nil
}

// PostClusterApply implements database.ClusterListener.
func (n *Notifier) PostClusterApply(leader bool, command database.Command, data interface{}, err error) {
	if !leader {
		return
	}

	switch data.(type) {
	case *Contact:
		c := data.(*Contact)

		cacheLock.Lock()
		if command == database.CommandSave {
			contactsCache[c.ID] = c
		} else if command == database.CommandDelete {
			delete(contactsCache, c.ID)
		}
		cacheLock.Unlock()

	case *ContactGroup:
		g := data.(*ContactGroup)

		cacheLock.Lock()
		if command == database.CommandSave {
			groupsCache[g.ID] = g
		} else if command == database.CommandDelete {
			delete(groupsCache, g.ID)
		}
		cacheLock.Unlock()

	case *eval.Evaluation:
		e := n.gotEvaluation(data.(*eval.Evaluation))
		if e != nil {
			logger.Red("notify", "%s", e.Error())
		}
	}
}

func (n *Notifier) gotEvaluation(e *eval.Evaluation) error {
	var check checks.Check
	err := n.db.One("ID", e.CheckID, &check)
	if err != nil {
		return err
	}

	state := e.History.Last(3).Reduce()

	if state == eval.StateUnknown {
		logger.Green("notify", "[%s] Ignoring %s %s", e.CheckID, state.String(), e.History.ColorString())
		return nil
	}

	duration := e.End.Sub(e.Start)

	stateCacheLock.RLock()
	lastState := stateCache[e.CheckID]
	stateCacheLock.RUnlock()

	if state == lastState {
		logger.Green("notify", "[%s] Ignoring unchanged state (Last: %s, Current: %s, Duration: %s) %s", e.CheckID, lastState, state, duration.String(), e.History.ColorString())
		// Nothing changed. Abort.
		return nil
	}

	stateCacheLock.Lock()
	stateCache[e.CheckID] = state
	stateCacheLock.Unlock()

	if lastState == eval.StateUnknown {
		logger.Green("notify", "[%s] Ignoring %s when previous state is %s %v", e.CheckID, state, lastState, e.History.ColorString())
		// Last state was unknown. Maybe we just started. Don't notify.
		return nil
	}

	if duration < check.Interval*2 && state == eval.StateDegraded {
		logger.Green("notify", "[%s] Ignoring %s for less than two cycles when degraded %s", e.CheckID, state, e.History.ColorString())
		return nil
	}

	logger.Yellow("notify", "%s is %s %s", e.CheckID, state, e.History.ColorString())

	if len(check.ContactGroups) == 0 {
		logger.Green("notify", "[%s] No-one to nofity, aborting", e.CheckID)
		return nil
	}

	text := fmt.Sprintf("%s is %s", e.CheckID, state.String())

	targetGroups := check.ContactGroups
	for _, groupID := range targetGroups {
		cacheLock.RLock()
		group, found := groupsCache[groupID]
		cacheLock.RUnlock()

		if !found {
			logger.Yellow("notify", "[%s] ContactGroup not found (%s)", e.CheckID, groupID)
			continue
		}

		// FIXME: Handle errors somehow.
		logger.Green("notify", "[%s] Notifying '%s'", e.CheckID, groupID)

		stats.CounterInc("notification_sent", 1)

		go group.Notify(text)
	}

	return nil
}
