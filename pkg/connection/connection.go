package connection

import (
	"io"
	"net"
)

// Connection is used to implement connection between narwhal server and client
//
// Serve: parse request method from tcp connection
// Close: close tcp connection
// Proxy: listen up binding port and proxy traffic
// ShouldProxy: connection use to auth and negotiate between client and server get false
// GetUID: get connection uuid
// SetAuthCtx: add authCtx to connection
// GetBindPort: get bind port of connection
// Reply: reply connection with reply code and payload
type Connection interface {
	Close()
	BindAndProxy(bPort int) error
	NewPConn(pConn net.Conn)
	SetAuthCtx(authCtx string)
	SetUID(uid string)
	SetToProxyConn()
	GetArrs() Arrs
}

func copyIO(srcConn, dstConn net.Conn) {
	defer srcConn.Close()
	defer dstConn.Close()
	io.Copy(dstConn, srcConn)
}
