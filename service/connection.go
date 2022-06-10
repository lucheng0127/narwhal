package service

import (
	"net"
	"sync"
)

var CM ConnManager

func init() {
	CM.ConnMap = make(map[uint16]*Connection)
	CM.LisMap = make(map[uint16]*Lister)
	CM.TransferConnMap = make(map[int]net.Conn)
	CM.ClientLocalPort = 0
}

type Connection struct {
	Conn net.Conn
	Key  uint16 // Key of connMap, set seq num as key
}

type Lister struct {
	Lister net.Listener
	Key    uint16 // Key of lisMap, set target port as key
}

type ConnManager struct {
	Mux     sync.Mutex
	ConnMap map[uint16]*Connection
	LisMap  map[uint16]*Lister
	// For server use the LocalAddr port as key when listen target port,
	// For server use the RemoteAddr port as key when listen server port,
	// For client use the RemoteAddr port as key
	TransferConnMap map[int]net.Conn
	ClientLocalPort int
}

// Functions of connManager
