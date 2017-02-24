package node

import (
	"bytes"
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"path"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hashicorp/raft"
	"github.com/hashicorp/raft-boltdb"

	"github.com/gansoi/gansoi/ca"
	"github.com/gansoi/gansoi/cluster"
	"github.com/gansoi/gansoi/database"
	"github.com/gansoi/gansoi/logger"
	"github.com/gansoi/gansoi/stats"
)

type (
	// Node represents a single gansoi node.
	Node struct {
		db            database.Database
		raft          *raft.Raft
		leader        bool
		leadersChans  []chan bool
		basePath      string
		listenersLock sync.RWMutex
		listeners     []database.Listener
		client        *http.Client
	}

	nodeInfo struct {
		Name    string            `json:"name" storm:"id"`
		Started time.Time         `json:"started"`
		Updated time.Time         `json:"updated"`
		Raft    map[string]string `json:"raft"`
	}
)

func init() {
	database.RegisterType(nodeInfo{})

	stats.CounterInit("apply_noleader")
	stats.CounterInit("apply_proxy")
	stats.CounterInit("apply_direct")
	stats.CounterInit("node_save")
	stats.CounterInit("node_one")
	stats.CounterInit("node_all")
	stats.CounterInit("node_delete")
}

// NewNode will initialize a new node.
func NewNode(stream *HTTPStream, datadir string, db database.Database, fsm raft.FSM, peers *cluster.Info, pair []tls.Certificate, coreCA *ca.CA) (*Node, error) {
	started := time.Now()

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			Certificates:       pair,
			RootCAs:            coreCA.CertPool(),
			InsecureSkipVerify: false,
		},
	}

	var err error
	n := &Node{
		db:     db,
		client: &http.Client{Transport: tr},
	}

	db.RegisterListener(n)

	// Raft config.
	conf := raft.DefaultConfig()
	conf.HeartbeatTimeout = 1000 * time.Millisecond
	conf.ElectionTimeout = 1000 * time.Millisecond
	conf.LeaderLeaseTimeout = 500 * time.Millisecond
	conf.CommitTimeout = 200 * time.Millisecond
	conf.Logger = logger.InfoLogger("raft")

	// If we have exactly one peer - and its ourself, we are bootstrapping.
	p, _ := peers.Peers()
	if len(p) == 1 && p[0] == peers.Self() {
		logger.Info("node", "Starting raft in bootstrap mode")
		conf.EnableSingleNode = true
		conf.DisableBootstrapAfterElect = false
	}

	transport := raft.NewNetworkTransportWithLogger(stream, 1, 0, logger.DebugLogger("raft-transport"))

	ss, err := raft.NewFileSnapshotStoreWithLogger(datadir, 5, logger.DebugLogger("raft-store"))
	if err != nil {
		return nil, err
	}

	store, err := raftboltdb.NewBoltStore(path.Join(datadir, "/raft.db"))
	if err != nil {
		return nil, err
	}

	n.raft, err = raft.NewRaft(
		conf,      // raft.Config
		fsm,       // raft.FSM
		store,     // raft.LogStore
		store,     // raft.StableStore
		ss,        // raft.SnapshotStore
		peers,     // raft.PeerStore
		transport, // raft.Transport
	)
	if err != nil {
		return nil, err
	}

	lch := n.raft.LeaderCh()

	// Let the cluster know how we're doing in two second intervals.
	tickChannel := time.NewTicker(time.Second * 2).C

	go func() {
		for {
			select {
			case leader := <-lch:
				n.leaderChange(leader)
			case <-tickChannel:
				var ni nodeInfo
				ni.Started = started
				ni.Updated = time.Now()
				ni.Name = peers.Self()
				ni.Raft = n.raft.Stats()

				n.Save(&ni)
			}
		}
	}()

	return n, nil
}

func (n *Node) leaderChange(leader bool) {
	n.leader = leader

	for _, ch := range n.leadersChans {
		ch <- leader
	}
}

// statsHandler will reply with a few raft statistics.
func (n *Node) statsHandler(c *gin.Context) {
	s := stats.GetAll()

	c.JSON(http.StatusOK, s)
}

// applyHandler can be used by other nodes to apply a log entry to the leader.
// The POST body should consists of the complete output from LogEntry.Byte().
func (n *Node) applyHandler(c *gin.Context) {
	if !n.leader {
		c.AbortWithStatus(http.StatusGone)
		return
	}

	defer c.Request.Body.Close()
	b, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	n.raft.Apply(b, time.Minute)
}

// nodesHandler will return stats for all nodes in the cluster.
func (n *Node) nodesHandler(c *gin.Context) {
	var all []nodeInfo

	err := n.db.All(&all, -1, 0, false)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, all)
}

// apply will apply the log entry to the local Raft node if it's leader, will
// forward to leader otherwise.
func (n *Node) apply(entry *database.LogEntry) error {
	// Only attempt this if the cluster is stable with a leader.
	if n.raft.Leader() == "" {
		stats.CounterInc("apply_noleader", 1)

		return raft.ErrLeader
	}

	if !n.leader {
		stats.CounterInc("apply_proxy", 1)

		r := bytes.NewReader(entry.Byte())
		l := n.raft.Leader()
		u := "https://" + l + n.basePath + "/apply"

		_, err := n.client.Post(u, "gansoi/entry", r)

		// FIXME: Implement some kind of retry logic here.

		return err
	}

	stats.CounterInc("apply_direct", 1)

	n.raft.Apply(entry.Byte(), time.Minute)

	return nil
}

// Save will save an object to the cluster database.
func (n *Node) Save(data interface{}) error {
	stats.CounterInc("node_save", 1)

	idsetter, ok := data.(database.IDSetter)
	if ok {
		idsetter.SetID()
	}

	entry := database.NewLogEntry(database.CommandSave, data)

	return n.apply(entry)
}

// One will retrieve one record from the cluster database.
func (n *Node) One(fieldName string, value interface{}, to interface{}) error {
	stats.CounterInc("node_one", 1)

	return n.db.One(fieldName, value, to)
}

// All lists all kinds of a type.
func (n *Node) All(to interface{}, limit int, skip int, reverse bool) error {
	stats.CounterInc("node_all", 1)

	return n.db.All(to, limit, skip, reverse)
}

// Delete deletes one record.
func (n *Node) Delete(data interface{}) error {
	stats.CounterInc("node_delete", 1)

	entry := database.NewLogEntry(database.CommandDelete, data)

	return n.apply(entry)
}

// RegisterListener will register a listener for new changes to the database.
func (n *Node) RegisterListener(listener database.Listener) {
	n.listenersLock.Lock()
	defer n.listenersLock.Unlock()

	n.listeners = append(n.listeners, listener)
}

// PostApply satisfies the database.Listener interface.
func (n *Node) PostApply(_ bool, command database.Command, data interface{}, err error) {
	n.listenersLock.RLock()
	defer n.listenersLock.RUnlock()

	for _, listener := range n.listeners {
		// We ignore the leader argument from caller. The caller is most likely
		// a local database that is unaware of raft leadership.
		go listener.PostApply(n.leader, command, data, err)
	}
}

// Router can be used to assign a Gin routergroup.
func (n *Node) Router(router *gin.RouterGroup) {
	n.basePath = router.BasePath()

	router.GET("/stats", n.statsHandler)
	router.GET("/nodes", n.nodesHandler)
	router.POST("/apply", n.applyHandler)
}

// AddPeer adds a new cluster/raft peer.
func (n *Node) AddPeer(name string) error {
	return n.raft.AddPeer(name).Error()
}

// LeaderCh is used to get a channel which delivers signals on acquiring or
// losing leadership. It sends true if we become the leader, and false if we
// lose it.
func (n *Node) LeaderCh() <-chan bool {
	ch := make(chan bool)

	n.leadersChans = append(n.leadersChans, ch)

	return ch
}
