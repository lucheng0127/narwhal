package connection

import (
	"context"
	"io"
	"net"

	"github.com/lucheng0127/narwhal/pkg/protocol"
)

// Connection is used to implement connection between narwhal server and client
// Server:
//
//	Auth - authentication the client
//	Bind - listen port and proxy to client authed
//	NewConn - handle new proxy connection
//	Serve - listen port that binded, when new connection establish,
//			get connection from connection channel, then proxy it
//	Proxy - use io.Copy(srcConn, dstConn) forward traffic between two tcp connection
//	Close - close connection
//	Notify - notify client this a new connection eastalished
//
// Client:
//
//	Auth - try to auth with uuid
//	Bind - tell server to bind a port
//	Serve - waiting for new connection connect to server bind port,
//	        then make a new connection to local port and a new connection to server
//	        then proxy to forward traffic between those two connection
//	Close - close connection
type Connection interface {
	Auth(ctx context.Context, pkg protocol.PKG)
	Bind(ctx context.Context, pkg protocol.PKG)
	NewConn(ctx context.Context, pkg protocol.PKG)
	Serve(ctx context.Context)
	Proxy()
	Close()
	ReplayWithCode(ctx context.Context, code byte) error
	ReplayWithAuthCtx(ctx context.Context, authCtx string) error
	ShouldProxy() bool
	SetAuthCtx(string)
	GetUID() string
	GetBindPort() int
}

func copyIO(srcConn, dstConn net.Conn) {
	defer srcConn.Close()
	defer dstConn.Close()
	io.Copy(dstConn, srcConn)
}
