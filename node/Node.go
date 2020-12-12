package node

import (
	"bytes"
	"crypto/tls"
	"errors"
	"expvar"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"sync"
	"syscall"
	"time"

	ginexpvar "github.com/gin-contrib/expvar"
	"github.com/gin-gonic/gin"
	hclog "github.com/hashicorp/go-hclog"
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"

	"github.com/gansoi/gansoi/ca"
	"github.com/gansoi/gansoi/database"
	"github.com/gansoi/gansoi/logger"
)

type (
	// Node represents a single gansoi node.
	Node struct {
		db               database.Reader
		raft             *raft.Raft
		leader           bool
		leadersChansLock sync.RWMutex
		leadersChans     []chan bool
		basePath         string
		listenersLock    sync.RWMutex
		listeners        []database.Listener
		client           *http.Client
		self             string
	}

	nodeInfo struct {
		Name    string            `json:"name" storm:"id"`
		Started time.Time         `json:"started"`
		Updated time.Time         `json:"updated"`
		Raft    map[string]string `json:"raft"`
	}
)

var (
	applyNoleader = expvar.NewInt("apply_noleader")
	applyProxy    = expvar.NewInt("apply_proxy")
	applyDirect   = expvar.NewInt("apply_direct")
	nodeSave      = expvar.NewInt("node_save")
	nodeOne       = expvar.NewInt("node_one")
	nodeAll       = expvar.NewInt("node_all")
	nodeFind      = expvar.NewInt("node_find")
	nodeDelete    = expvar.NewInt("node_delete")

	// ErrNoLeader will be returned if an operation requires a leader - but
	// we have none.
	ErrNoLeader = errors.New("no leader")
)

func init() {
	database.RegisterType(nodeInfo{})
}

// NewNode will initialize a new node.
func NewNode(stream raft.StreamLayer, datadir string, db database.Reader, fsm raft.FSM, self string, pair []tls.Certificate, coreCA *ca.CA) (*Node, func() error, error) {
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
		self:   self,
	}

	// Raft config.
	conf := raft.DefaultConfig()
	conf.HeartbeatTimeout = 1000 * time.Millisecond
	conf.ElectionTimeout = 1000 * time.Millisecond
	conf.LeaderLeaseTimeout = 500 * time.Millisecond
	conf.CommitTimeout = 200 * time.Millisecond
	conf.Logger = hclog.New(nil)
	conf.SnapshotInterval = time.Second * 60
	conf.SnapshotThreshold = 100
	conf.LocalID = raft.ServerID(n.self)
	conf.ProtocolVersion = 3

	transport := raft.NewNetworkTransportWithLogger(stream, 1, 0, nil)

	ss, err := raft.NewFileSnapshotStoreWithLogger(datadir, 2, nil)
	if err != nil {
		return nil, nil, err
	}

	raftDBPath := path.Join(datadir, "/raft.db")
	store, err := raftboltdb.NewBoltStore(raftDBPath)
	if err != nil {
		return nil, nil, err
	}

	var st syscall.Stat_t

	if syscall.Stat(path.Dir(raftDBPath), &st) == nil {
		// Try to fix ownership, fail silently.
		os.Chown(raftDBPath, int(st.Uid), int(st.Gid))

		// "snapshots" is raft.snapPath.
		os.Chown(path.Join(datadir, "snapshots"), int(st.Uid), int(st.Gid))
	}

	n.raft, err = raft.NewRaft(
		conf,      // raft.Config
		fsm,       // raft.FSM
		store,     // raft.LogStore
		store,     // raft.StableStore
		ss,        // raft.SnapshotStore
		transport, // raft.Transport
	)
	if err != nil {
		return nil, nil, err
	}

	return n, func() error {
		err := n.raft.Shutdown().Error()
		if err != nil {
			return err
		}

		err = store.Close()
		if err != nil {
			return err
		}

		return transport.Close()
	}, nil
}

// Run will start a few background tasks needed for the node.
func (n *Node) Run() {
	started := time.Now()

	lch := n.raft.LeaderCh()

	// Let the cluster know how we're doing in two second intervals.
	tickChannel := time.NewTicker(time.Second * 2).C

	go func() {
		for {
			select {
			case leader := <-lch:
				n.leaderChange(leader)
			case <-tickChannel:
				// Only try to save if we have a leader
				if n.raft.Leader() == "" {
					break
				}

				var ni nodeInfo
				ni.Started = started
				ni.Updated = time.Now()
				ni.Name = n.self
				ni.Raft = n.raft.Stats()

				err := n.Save(&ni)
				if err != nil {
					logger.Info("node", "Error from cluster save: %s", err.Error())
				}
			}
		}
	}()
}

// Bootstrap will start a new Gansoi cluster. This should only be called for
// new clusters. Not new nodes in a existing cluster.
func (n *Node) Bootstrap() error {
	configuration := raft.Configuration{
		Servers: []raft.Server{{
			ID:       raft.ServerID(n.self),
			Address:  raft.ServerAddress(n.self),
			Suffrage: raft.Voter,
		}},
	}

	return n.raft.BootstrapCluster(configuration).Error()
}

func (n *Node) leaderChange(leader bool) {
	n.leader = leader

	n.leadersChansLock.RLock()
	for _, ch := range n.leadersChans {
		ch <- leader
	}
	n.leadersChansLock.RUnlock()
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
		applyNoleader.Add(1)

		return ErrNoLeader
	}

	if !n.leader {
		applyProxy.Add(1)

		r := bytes.NewReader(entry.Byte())
		l := n.raft.Leader()
		u := "https://" + string(l) + n.basePath + "/apply"

		_, err := n.client.Post(u, "gansoi/entry", r)

		// FIXME: Implement some kind of retry logic here.

		return err
	}

	applyDirect.Add(1)

	n.raft.Apply(entry.Byte(), time.Minute)

	return nil
}

// Save will save an object to the cluster database.
func (n *Node) Save(data interface{}) error {
	nodeSave.Add(1)

	idsetter, ok := data.(database.IDSetter)
	if ok {
		idsetter.SetID()
	}

	entry := database.NewLogEntry(database.CommandSave, data)

	return n.apply(entry)
}

// One will retrieve one record from the cluster database.
func (n *Node) One(fieldName string, value interface{}, to interface{}) error {
	nodeOne.Add(1)

	return n.db.One(fieldName, value, to)
}

// All lists all kinds of a type.
func (n *Node) All(to interface{}, limit int, skip int, reverse bool) error {
	nodeAll.Add(1)

	return n.db.All(to, limit, skip, reverse)
}

// Find find objects of type.
func (n *Node) Find(field string, value interface{}, to interface{}, limit int, skip int, reverse bool) error {
	nodeFind.Add(1)

	return n.db.Find(field, value, to, limit, skip, reverse)
}

// Delete deletes one record.
func (n *Node) Delete(data interface{}) error {
	nodeDelete.Add(1)

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
func (n *Node) PostApply(_ bool, command database.Command, data interface{}) {
	n.listenersLock.RLock()
	defer n.listenersLock.RUnlock()

	for _, listener := range n.listeners {
		// We ignore the leader argument from caller. The caller is most likely
		// a local database that is unaware of raft leadership.
		go listener.PostApply(n.leader, command, data)
	}
}

// Router can be used to assign a Gin routergroup.
func (n *Node) Router(router *gin.RouterGroup) {
	n.basePath = router.BasePath()

	router.GET("/stats", ginexpvar.Handler())
	router.GET("/nodes", n.nodesHandler)
	router.POST("/apply", n.applyHandler)
}

// AddPeer adds a new cluster/raft peer.
func (n *Node) AddPeer(name string) error {
	return n.raft.AddVoter(raft.ServerID(name), raft.ServerAddress(name), 0, 0).Error()
}

// LeaderCh is used to get a channel which delivers signals on acquiring or
// losing leadership. It sends true if we become the leader, and false if we
// lose it.
func (n *Node) LeaderCh() <-chan bool {
	ch := make(chan bool)

	n.leadersChansLock.Lock()
	n.leadersChans = append(n.leadersChans, ch)
	n.leadersChansLock.Unlock()

	return ch
}

// LastIndex returns the last Raft index in stable storage, either from the
// last log or from the last snapshot.
func (n *Node) LastIndex() uint64 {
	return n.raft.LastIndex()
}
