package connection

import (
	"io"
	"net"
)

// Connection is used to implement connection between narwhal server and client
// Server:
//
//		Auth - authentication the client
//		Bind - listen port and proxy to client authed
//		Serve - listen port that binded, when new connection establish,
//				get connection from connection channel, then proxy it
//		Proxy - use io.Copy(srcConn, dstConn) forward traffic between two tcp connection
//		Close - close connection
//	 	Notify - notify client this a new connection eastalished
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
	Auth()
	Serve()
	Proxy()
	Close()
	ReplayWithCode(code byte)
	ReplayWithAuthCtx()
}

func copyIO(srcConn, dstConn net.Conn) {
	defer srcConn.Close()
	defer dstConn.Close()
	io.Copy(dstConn, srcConn)
}
