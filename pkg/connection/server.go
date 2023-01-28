package connection

import (
	"fmt"
	"net"
	"runtime/debug"

	logger "github.com/lucheng0127/narwhal/internal/pkg/log"
	"github.com/lucheng0127/narwhal/internal/pkg/utils"
	"github.com/lucheng0127/narwhal/pkg/protocol"
)

type SConn struct {
	UID         string
	AuthCtx     string
	BindPort    int
	Ready       chan bool    // After Receive
	ln          net.Listener // Listener of bind port
	conn        net.Conn
	proxyConnCh chan net.Conn // Connection used to port forwarding
}

func NewServerConnection(conn net.Conn) *SConn {
	return &SConn{
		conn:        conn,
		BindPort:    -1,
		UID:         "",
		Ready:       make(chan bool, 1),
		proxyConnCh: make(chan net.Conn, 1)}
}

func (c *SConn) Auth() {

}

func (c *SConn) Notify() {
	// TODO(shawnlu): Implement it
}

func (c *SConn) Close() {
	defer c.conn.Close()
	c.ReplayWithCode(protocol.RepConnClose)
}

func (c *SConn) ReplayWithCode(code byte) {
	// TODO(shawnlu): Send close connection to connection
}

func (c *SConn) ReplayWithAuthCtx() {}

func (c *SConn) Serve() {}

func (c *SConn) forwarding(sConn, tConn net.Conn) {
	defer func() {
		if r := recover(); r != nil {
			sConn.Close()
			tConn.Close()
			ctx := utils.NewTraceContext()
			logger.Warn(ctx, fmt.Sprintf("Proxy %s %s end, because of %s", sConn.RemoteAddr().String(), tConn.RemoteAddr().String(), debug.Stack()))
		}
	}()

	ctx := utils.NewTraceContext()
	logger.Debug(ctx, fmt.Sprintf("Proxy %s %s\n", sConn.RemoteAddr().String(), tConn.RemoteAddr().String()))
	go copyIO(sConn, tConn)
	go copyIO(tConn, sConn)
}

func (c *SConn) Proxy() {
	defer c.ln.Close()
	ctx := utils.NewTraceContext()
	logger.Info(ctx, fmt.Sprintf("Start to serve %s\n", c.ln.Addr().String()))

	for {
		conn, err := c.ln.Accept()
		if err != nil {
			logger.Error(ctx, err.Error())
			continue
		}

		// Tell client this a new connection establised, client will
		c.Notify()

		tConn := <-c.proxyConnCh
		go c.forwarding(conn, tConn)
	}
}
