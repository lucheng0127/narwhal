package connection

import (
	"fmt"
	"io"
	"net"
	"runtime/debug"

	logger "github.com/lucheng0127/narwhal/internal/pkg/log"
	"github.com/lucheng0127/narwhal/internal/pkg/utils"
)

type Arrs struct {
	UID         string
	AuthCtx     string
	BindPort    int
	ln          net.Listener // Listener of bind port
	Conn        net.Conn
	ProxyConnCh chan net.Conn // Connection used to port forwarding
	ProxyConn   bool
}

type Client interface {
	Auth(uid string) error
	Bind(rPort uint16) error
	MonitorAndProxy(lPort uint16) error
	Close()
}

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

func ioSwitch(pConn, tConn net.Conn) {
	defer func() {
		if r := recover(); r != nil {
			pConn.Close()
			tConn.Close()
			ctx := utils.NewTraceContext()
			logger.Warn(ctx, fmt.Sprintf("stop proxy %s %s\n%s", pConn.RemoteAddr().String(), tConn.RemoteAddr().String(), debug.Stack()))
		}
	}()

	ctx := utils.NewTraceContext()
	logger.Debug(ctx, fmt.Sprintf("proxy %s %s\n", pConn.RemoteAddr().String(), tConn.RemoteAddr().String()))
	go copyIO(pConn, tConn)
	go copyIO(tConn, pConn)
}
