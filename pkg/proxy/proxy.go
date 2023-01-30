package proxy

import (
	"net"

	"github.com/lucheng0127/narwhal/pkg/connection"
)

type Server interface {
	Launch() error
	Serve(net.Listener) error
	ServeConn(conn connection.Connection)
	Stop()
}
