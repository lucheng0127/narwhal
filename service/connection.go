package service

import (
	"net"
	"sync"
)

const (
	S_DUBIOUS uint8 = 0xa0
	S_READY   uint8 = 0xa1
	S_CLOSED  uint8 = 0xa2
)

type connection struct {
	conn   net.Conn
	status uint8
}

type lister struct {
	lister net.Listener
	status uint8
}

type connManager struct {
	mux             sync.Mutex
	connMap         map[string]*connection // Socket address as key, store net.Conn
	lisMap          map[int]*lister        // Port as key, store TCP lister
	transferConnKey string                 // Key of connection, used to forward socket traffic
}
