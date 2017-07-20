package ssh

import (
	"sync"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/gansoi/gansoi/logger"
)

type (
	connection struct {
		connectLock sync.Mutex
		lastUse     time.Time
		client      *ssh.Client
		refCount    int
	}
)

var (
	poolLock sync.Mutex
	pool     = make(map[SSH]*connection)

	closeAfter = time.Second * 30
)

func init() {
	go loop()
}

func loop() {
	ticker := time.Tick(closeAfter / 10)

	for t := range ticker {
		poolLock.Lock()

		for s, conn := range pool {
			conn.connectLock.Lock()
			if t.Sub(conn.lastUse) > closeAfter && conn.refCount == 0 && conn.client != nil {
				conn.client.Close()
				conn.client = nil
				logger.Debug("ssh", "Closing unused connection %s", s.Address)
			}
			conn.connectLock.Unlock()
		}

		poolLock.Unlock()
	}
}

func connect(s SSH) (*ssh.Client, error) {
	poolLock.Lock()
	conn, found := pool[s]
	if !found {
		conn = &connection{}

		pool[s] = conn
	}

	conn.connectLock.Lock()
	defer conn.connectLock.Unlock()

	poolLock.Unlock()

	conn.lastUse = time.Now()

	if conn.client != nil {
		conn.refCount++

		return conn.client, nil
	}

	client, err := s.connect()
	if err != nil {
		return nil, err
	}

	conn.refCount++
	conn.client = client

	return client, nil
}

// done returns a SSH connection to the pool.
func done(s SSH) {
	poolLock.Lock()
	defer poolLock.Unlock()

	conn, found := pool[s]
	if found && conn.client != nil {
		conn.lastUse = time.Now()
		conn.refCount--
	}

	if !found {
		logger.Info("ssh", "Connection not found in pool when returning (%s)", s)
	}
}
