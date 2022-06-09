package service

import (
	"net"
	"sync"
)

var CM ConnManager

func init() {
	CM.ConnMap = make(map[uint16]*Connection)
	CM.LisMap = make(map[uint16]*Lister)
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
	Mux             sync.Mutex
	ConnMap         map[uint16]*Connection
	LisMap          map[uint16]*Lister
	TransferConnMap map[int]net.Conn // For server use the LocalAddr port as key, for client use the Remote Addr port as key
}

// Functions of connManager
