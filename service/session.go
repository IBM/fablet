package service

import (
	"sync"
	"time"

	"github.com/IBM/fablet/api"
)

const (
	// ConnMonitorInterval  to check the connection every interval.
	ConnMonitorInterval = time.Second * 30
	// ConnInactiveLongest  the connection will be closed and removed if be not actived for longer than the inactive time.
	ConnInactiveLongest = time.Minute * 10
)

// ConnSession to store all connections.
// TODO to have NetworkConnection with lock. ***
type ConnSession struct {
	Connections map[string]*api.NetworkConnection
	sync.RWMutex
}

// FindConn find connection from session.
func (connSession *ConnSession) findConn(id string) (*api.NetworkConnection, bool) {
	conn, ok := connSession.Connections[id]
	return conn, ok
}

func (connSession *ConnSession) storeConn(conn *api.NetworkConnection) {
	connSession.Lock()
	defer connSession.Unlock()

	// Double check, to avoid concurrent storing.
	// This might cause exception if the tmpConn is being used, but it is not critical.
	// TODO
	if tmpConn, ok := connSession.findConn(conn.Identifier); ok {
		logger.Debugf("Double check and found stored connection of %s.", conn.Identifier)
		tmpConn.Close()
	}
	connSession.Connections[conn.Identifier] = conn
}

func (connSession *ConnSession) removeConn(id string) {
	connSession.Lock()
	defer connSession.Unlock()
	if conn, ok := connSession.findConn(id); ok {
		// TODO It might be not safe here, the conn might be in using.
		conn.Close()
		delete(connSession.Connections, id)
	}
}

// A globla variable.
var connSession *ConnSession

// Init of package level.
func init() {
	logger.Info("Initialize the connection store.")
	connSession = &ConnSession{
		Connections: make(map[string]*api.NetworkConnection),
	}
	go monitorConnSession()
}

func monitorConnSession() {
	tk := time.NewTicker(ConnMonitorInterval)
	for t := range tk.C {
		for id, conn := range connSession.Connections {
			afterActive := time.Since(conn.ActiveTime)
			if afterActive > ConnInactiveLongest {
				logger.Infof("Connection %s after active time: %v (now %v), so then to be closed and removed.", id, afterActive, t)
				go connSession.removeConn(id)
			} else {
				logger.Infof("Connection %s after active time: %v.", id, afterActive)
			}
		}
	}
}

// GetConnection get connection from session, might be existing or new.
func GetConnection(connProfile *api.ConnectionProfile, participant *api.Participant, useDiscovery bool) (*api.NetworkConnection, error) {
	id := string(api.CalConnIdentifier(connProfile, participant, useDiscovery))
	if conn, ok := connSession.findConn(id); ok {
		logger.Debugf("Find stored connection of %s.", id)
		conn.ActiveTime = time.Now()
		return conn, nil
	}

	return NewConnection(connProfile, participant, useDiscovery)
}

// NewConnection create connection and then store it into session.
func NewConnection(connProfile *api.ConnectionProfile, participant *api.Participant, useDiscovery bool) (*api.NetworkConnection, error) {
	conn, err := api.NewConnection(connProfile, participant, useDiscovery)
	if err == nil {
		logger.Debugf("Store new connection of %s.", conn.Identifier)
		// Might be discarded in storeConn function, but not critical.
		go connSession.storeConn(conn)
		return conn, nil
	}
	return nil, err
}
