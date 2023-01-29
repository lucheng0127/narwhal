package connection

import (
	"context"
	"encoding/binary"
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
	ProxyConnCh chan net.Conn // Connection used to port forwarding
	ProxyConn   bool
}

func NewServerConnection(conn net.Conn) *SConn {
	return &SConn{
		conn:        conn,
		BindPort:    -1,
		UID:         "",
		Ready:       make(chan bool, 1),
		ProxyConnCh: make(chan net.Conn, 1),
		ProxyConn:   false}
}

func (c *SConn) Auth(ctx context.Context, pkg protocol.PKG) {
	// Set payload to UID
	c.UID = string(pkg.GetPayload())
}

func (c *SConn) Bind(ctx context.Context, pkg protocol.PKG) {
	// Set payload to BindPort
	c.BindPort = int(binary.BigEndian.Uint16(pkg.GetPayload()))

	// Send true to Ready in the last
	c.Ready <- true
}

func (c *SConn) NewConn(ctx context.Context, pkg protocol.PKG) {
	c.AuthCtx = string(pkg.GetPayload())

	c.ProxyConn = true
	// Send true to Ready in the last
	c.Ready <- true
}

func (c *SConn) GetUID() string {
	return c.UID
}

func (c *SConn) SetAuthCtx(authCtx string) {
	c.AuthCtx = authCtx
}

func (c *SConn) GetBindPort() int {
	return c.BindPort
}

func (c *SConn) Notify() {
	// TODO(shawnlu): Implement it
}

func (c *SConn) Close() {
	defer c.conn.Close()
	c.ReplayWithCode(protocol.RepConnClose)
}

func (c *SConn) ShouldProxy() bool {
	return !c.ProxyConn
}

func (c *SConn) ReplayWithCode(code byte) {
	// TODO(shawnlu): Send close connection to connection
}

func (c *SConn) ReplayWithAuthCtx() {
	// TODO(shawnlu): Implement it
}

func (c *SConn) Serve(ctx context.Context) {
	for {
		if c.ProxyConn {
			// ProxyConn use to proxy do not serve
			break
		}

		// Parse request method
		var pkg protocol.PKG = protocol.NewRequestMethod()
		err := pkg.Parse(ctx, c.conn)
		if err != nil {
			logger.Error(ctx, err.Error())
		}

		// Handle request method, after auth and bind, send true to channel ready
		c.HandleMethod(ctx, pkg)
	}
}

// When connection cmd is CmdNewConn, it means this connection no need auth
// it's a proxy connection, should end of connection.Serve
func (c *SConn) HandleMethod(ctx context.Context, pkg protocol.PKG) {

	cmd := pkg.GetCmd()
	switch cmd {
	case protocol.CmdAuth:
		c.Auth(ctx, pkg)
	case protocol.CmdBind:
		c.Bind(ctx, pkg)
	case protocol.CmdClose:
		c.conn.Close()
	case protocol.CmdNewConn:
		c.NewConn(ctx, pkg)
	}
}

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
	if c.ln == nil {
		return
	}

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

		tConn := <-c.ProxyConnCh
		go c.forwarding(conn, tConn)
	}
}
