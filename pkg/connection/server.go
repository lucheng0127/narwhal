package connection

import (
	"fmt"
	"net"
)

type Arrs struct {
	UID         string
	AuthCtx     string
	BindPort    int
	Ready       chan bool    // After Receive
	ln          net.Listener // Listener of bind port
	Conn        net.Conn
	ProxyConnCh chan net.Conn // Connection used to port forwarding
	ProxyConn   bool
}
type SConn struct {
	arrs Arrs
}

func NewServerConnection(conn net.Conn) Connection {
	c := new(SConn)
	c.arrs.Conn = conn
	return c
}

func (c *SConn) SetToProxyConn() {
	c.arrs.ProxyConn = true
}

func (c *SConn) SetAuthCtx(authCtx string) {
	c.arrs.AuthCtx = authCtx
}

func (c *SConn) SetUID(uid string) {
	c.arrs.UID = uid
}

func (c *SConn) NewPConn(conn net.Conn) {
	c.arrs.ProxyConnCh <- conn
}

func (c *SConn) Close() {
	c.arrs.Conn.Close()
}

func (c *SConn) GetArrs() Arrs {
	return c.arrs
}

func (c *SConn) BindAndProxy(bPort int) error {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", c.arrs.BindPort))
	if err != nil {
		return err
	}
	c.arrs.ln = ln

	// TODO
	return nil
}

//
//func (c *SConn) forwarding(sConn, tConn net.Conn) {
//	defer func() {
//		if r := recover(); r != nil {
//			sConn.Close()
//			tConn.Close()
//			ctx := utils.NewTraceContext()
//			logger.Warn(ctx, fmt.Sprintf("Proxy %s %s end, because of %s", sConn.RemoteAddr().String(), tConn.RemoteAddr().String(), debug.Stack()))
//		}
//	}()
//
//	ctx := utils.NewTraceContext()
//	logger.Debug(ctx, fmt.Sprintf("Proxy %s %s\n", sConn.RemoteAddr().String(), tConn.RemoteAddr().String()))
//	go copyIO(sConn, tConn)
//	go copyIO(tConn, sConn)
//}
//
//func (c *SConn) Proxy() {
//	if c.ln == nil {
//		return
//	}
//
//	defer c.ln.Close()
//	ctx := utils.NewTraceContext()
//	logger.Info(ctx, fmt.Sprintf("Start to serve %s\n", c.ln.Addr().String()))
//
//	for {
//		conn, err := c.ln.Accept()
//		if err != nil {
//			logger.Error(ctx, err.Error())
//			continue
//		}
//
//		// Tell client this a new connection establised, client will
//		c.Notify()
//
//		tConn := <-c.ProxyConnCh
//		go c.forwarding(conn, tConn)
//	}
//}
