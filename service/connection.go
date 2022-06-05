package service

import (
	"net"
	"sync"
)

const (
	// Connection status
	CONN_INIT     uint8 = 0x01
	CONN_PENDING  uint8 = 0x02
	CONN_READY    uint8 = 0x04
	CONN_UNHEALTH uint8 = 0x08

	// Lister status
	LIS_READY    uint8 = 0xf0
	LIS_UNHEALTH uint8 = 0xf1
)

type connection struct {
	conn   net.Conn
	status uint8
}

type lister struct {
	listen net.Listener
	status uint8
}

type connPeer struct {
	local  *lister
	remote *connection
}

type connManager struct {
	mux     sync.Mutex
	connMap map[int]*connPeer
}
