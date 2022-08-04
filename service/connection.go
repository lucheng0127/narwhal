package service

import (
	"net"
	"sync"
)

const (
	C_ACTIVE  string = "Active"
	C_CLOSED  string = "Closed"
	C_PENDING string = "Pending"
)

var CM ConnManager

func init() {
	CM.ConnMap = make(map[uint16]*Connection)
	CM.LisMap = make(map[uint16]*Lister)
	CM.TConnMap = make(map[string]net.Conn)
	CM.ClientLocalPort = 0
}

type Connection struct {
	Conn   net.Conn
	Key    uint16 // Key of connMap, set seq num as key
	Status string
}

type Lister struct {
	Lister net.Listener
	Key    uint16 // Key of lisMap, set target port as key
}

type ConnManager struct {
	Mux     sync.RWMutex
	ConnMap map[uint16]*Connection
	LisMap  map[uint16]*Lister
	// For server use the LocalAddr port as key when listen target port,
	// For server use the RemoteAddr port as key when listen server port,
	// For client use the RemoteAddr port as key
	TConnMap        map[string]net.Conn
	ClientLocalPort int
	ServerAddr      string
}

// Functions of connManager
