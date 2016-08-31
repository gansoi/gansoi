package node

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hashicorp/raft"

	"github.com/abrander/gansoi/database"
)

type (
	// Node represents a single gansoi node.
	Node struct {
		db       *database.Database
		peers    *PeerStore
		raft     *raft.Raft
		leader   bool
		stream   *HTTPStream
		basePath string
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
}

// NewNode will initialize a new node.
func NewNode(secret string, db *database.Database, peerStore *PeerStore) (*Node, error) {
	started := time.Now()

	var err error
	n := &Node{
		db:    db,
		peers: peerStore,
	}

	// Raft config.
	conf := raft.DefaultConfig()
	conf.HeartbeatTimeout = 1000 * time.Millisecond
	conf.ElectionTimeout = 1000 * time.Millisecond
	conf.LeaderLeaseTimeout = 500 * time.Millisecond
	conf.CommitTimeout = 200 * time.Millisecond
	conf.Logger = log.New(os.Stdout, "      RAFT ", log.Lmicroseconds|log.Lshortfile)

	// Set up nice HTTP based transport.
	n.stream, err = NewHTTPStream(peerStore.Self(), secret)
	if err != nil {
		panic(err)
	}

	logger := log.New(os.Stdout, " TRANSPORT ", log.Ldate|log.Lmicroseconds|log.Lshortfile)
	transport := raft.NewNetworkTransportWithLogger(n.stream, 1, 0, logger)

	logger = log.New(os.Stdout, " SNAPSTORE ", log.Ldate|log.Lmicroseconds|log.Lshortfile)
	ss, err := raft.NewFileSnapshotStoreWithLogger("/tmp/"+peerStore.Self(), 5, logger)
	if err != nil {
		panic(err)
	}

	n.raft, err = raft.NewRaft(
		conf,                 // raft.Config
		n.db,                 // raft.FSM
		raft.NewInmemStore(), // raft.LogStore
		raft.NewInmemStore(), // raft.StableStore
		ss,                   // raft.SnapshotStore
		n.peers,              // raft.PeerStore
		transport,            // raft.Transport
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
				n.leader = leader
			case <-tickChannel:
				var ni nodeInfo
				ni.Started = started
				ni.Updated = time.Now()
				ni.Name = peerStore.Self()
				ni.Raft = n.raft.Stats()

				n.Save(&ni)
			}
		}
	}()

	return n, nil
}

// raftHandler is a handler for the other nodes.
func (n *Node) raftHandler(c *gin.Context) {
	n.stream.ServeHTTP(c.Writer, c.Request)
}

// statsHandler will reply with a few raft statistics.
func (n *Node) statsHandler(c *gin.Context) {
	c.JSON(http.StatusOK, n.raft.Stats())
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
		return raft.ErrLeader
	}

	if !n.leader {
		r := bytes.NewReader(entry.Byte())
		l := n.raft.Leader()
		u := "https://" + l + n.basePath + "/apply"

		_, err := http.Post(u, "gansoi/entry", r)

		return err
	}

	n.raft.Apply(entry.Byte(), time.Minute)

	return nil
}

// Save will save an object to the cluster database.
func (n *Node) Save(data interface{}) error {
	entry := database.NewLogEntry(database.CommandSave, data)

	return n.apply(entry)
}

// One will retrieve one record from the cluster database.
func (n *Node) One(fieldName string, value interface{}, to interface{}) error {
	return n.db.One(fieldName, value, to)
}

// All lists all kinds of a type.
func (n *Node) All(to interface{}, limit int, skip int, reverse bool) error {
	return n.db.All(to, limit, skip, reverse)
}

// Delete deletes one record.
func (n *Node) Delete(data interface{}) error {
	entry := database.NewLogEntry(database.CommandDelete, data)

	return n.apply(entry)
}

func (n *Node) resultHandler(c *gin.Context) {
	checkResult := &database.CheckResult{}

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

	err = json.Unmarshal(b, checkResult)
	if err == nil {
		n.processResult(checkResult)
	}
}

func (n *Node) processResult(checkResult *database.CheckResult) {
	if !n.leader {
		return
	}

	fmt.Printf("Results are in: %+v\n", checkResult)

	// FIXME: Evaluate these somehow.
}

// SubmitResult will submit a new result set to the cluster for further processing.
func (n *Node) SubmitResult(checkID string, err error, result interface{}) error {
	checkResult := &database.CheckResult{
		CheckID:   checkID,
		Node:      n.peers.Self(),
		TimeStamp: time.Now(),
		Results:   result,
	}

	if err != nil {
		checkResult.Error = err.Error()
	}

	b, _ := json.Marshal(checkResult)

	if !n.leader {
		r := bytes.NewReader(b)
		l := n.raft.Leader()
		u := "https://" + l + n.basePath + "/result"

		_, e := http.Post(u, "gansoi/result", r)

		return e
	}

	// We JSON marshal/unmarshall everything to make sure processResult always
	// see the same data from followers AND leader. We assume this never fails.
	checkResult2 := &database.CheckResult{}
	json.Unmarshal(b, checkResult2)

	n.processResult(checkResult2)

	return nil
}

// Router can be used to assign a Gin routergroup.
func (n *Node) Router(router *gin.RouterGroup) {
	n.basePath = router.BasePath()

	router.GET("", n.raftHandler)
	router.GET("/stats", n.statsHandler)
	router.GET("/nodes", n.nodesHandler)
	router.POST("/apply", n.applyHandler)
	router.POST("/result", n.resultHandler)
}
