package connection

import "net"

type CConn struct {
	arrs Arrs
}

func NewClient(conn net.Conn) Client {
	c := new(CConn)
	c.arrs.Conn = conn
	return c
}

// Auth connection with uid, get authCtx from reply
func (c *CConn) Auth(uid string) error {
	c.arrs.UID = uid
	return nil
}

// Send ReqBind with payload rPort
func (c *CConn) Bind(rPort uint16) error {
	return nil
}

// Monitor notify and start proxy
func (c *CConn) MonitorAndProxy(lPort uint16) error {
	return nil
}

func (c *CConn) Close() {
	c.arrs.Conn.Close()
}
