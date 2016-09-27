package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"sync"
	"time"

	"golang.org/x/net/websocket"

	"github.com/abrander/gansoi/database"
)

type (
	// Live takes care of informing websocket clients about database changes.
	Live struct {
		lock      sync.RWMutex
		listeners map[*websocket.Conn]bool
	}

	// message describes the wire protocol.
	message struct {
		Command string      `json:"command"`
		Type    string      `json:"type"`
		Data    interface{} `json:"data"`
	}
)

// NewLive starts a new live handler.
func NewLive() *Live {
	return &Live{
		listeners: make(map[*websocket.Conn]bool),
	}
}

// PostClusterApply implements node.Listener.
func (l *Live) PostClusterApply(leader bool, command database.Command, data interface{}, err error) {
	// Interesting events for now.
	m := map[string]string{
		"*checks.Check":       "check",
		"*checks.CheckResult": "checkresult",
		"*node.nodeInfo":      "nodeinfo",
	}

	// Get the type as a string to avoid importing the generating package.
	typ := reflect.TypeOf(data).String()

	typ, interesting := m[typ]

	if interesting {
		payload := message{
			Command: command.String(),
			Type:    typ,
			Data:    data,
		}

		b, _ := json.Marshal(payload)

		l.lock.RLock()
		for listener := range l.listeners {
			// FIXME: This will block if a single client is unable to receive fast enough.
			_, err := listener.Write(b)
			if err != nil {
				// FIXME: Handle errors somehow.
				fmt.Printf("err: %s\n", err.Error())
			}
		}
		l.lock.RUnlock()
	}
}

// handleWS will handle incoming websocket connections.
func (l *Live) handleWS(conn *websocket.Conn) {
	// Undo deadlines set by webserver.
	conn.SetDeadline(time.Time{})
	conn.SetWriteDeadline(time.Time{})
	conn.SetReadDeadline(time.Time{})

	l.lock.Lock()
	l.listeners[conn] = true
	l.lock.Unlock()

	buf := make([]byte, 1024)

	// If read returns it means that the client is disconnected (when returning
	// an error), or that the client sent of something which is a protocol
	// violation. We end the connection. Goodbye.
	conn.Read(buf)

	l.lock.Lock()
	delete(l.listeners, conn)
	l.lock.Unlock()
}

// ServeHTTP implements http.Handler.
func (l *Live) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	handler := websocket.Handler(l.handleWS)

	handler.ServeHTTP(w, req)
}
