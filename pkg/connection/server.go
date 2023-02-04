package connection

import (
	"fmt"
	"net"

	logger "github.com/lucheng0127/narwhal/internal/pkg/log"
	"github.com/lucheng0127/narwhal/internal/pkg/utils"
	"github.com/lucheng0127/narwhal/pkg/protocol"
)

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

func (c *SConn) notify() error {
	// Notify client new connection establish with authCtx through c.Conn
	pkt := protocol.NewPkt(protocol.RepNotify, []byte(c.arrs.AuthCtx))
	pktData, err := pkt.Encode()
	if err != nil {
		return err
	}

	n, err := c.arrs.Conn.Write(pktData)
	if err != nil {
		return err
	}
	ctx := utils.NewTraceContext()
	logger.Debug(ctx, fmt.Sprintf("send [%d] bytes notify data to connection [%s]", n, c.arrs.Conn.RemoteAddr().String()))
	return nil
}

func (c *SConn) BindAndProxy(bPort int) error {
	ctx := utils.NewTraceContext()
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", bPort))
	if err != nil {
		return err
	}
	c.arrs.ln = ln

	for {
		conn, err := ln.Accept()
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("connection establish with port [%d] %s", bPort, err.Error()))
		}

		err = c.notify()
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("send notify to connection [%s] %s", c.arrs.Conn.RemoteAddr().String(), err.Error()))
			conn.Close()
			continue
		}

		tConn := <-c.arrs.ProxyConnCh
		go ioSwitch(conn, tConn)
	}
}
