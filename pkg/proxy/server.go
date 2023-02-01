package proxy

import (
	"fmt"
	"net"
	"runtime/debug"

	logger "github.com/lucheng0127/narwhal/internal/pkg/log"
	"github.com/lucheng0127/narwhal/internal/pkg/utils"
	"github.com/lucheng0127/narwhal/pkg/connection"
	"github.com/lucheng0127/narwhal/pkg/protocol"
	uuid "github.com/satori/go.uuid"
)

type ProxyServer struct {
	port       int // Service port
	ln         net.Listener
	users      map[string]string
	authedConn map[string]connection.Connection
}

func NewProxyServer(opts ...Option) Server {
	s := new(ProxyServer)
	for _, o := range opts {
		o(s)
	}

	ctx := utils.NewTraceContext()
	if s.port == 0 {
		logger.Warn(ctx, fmt.Sprintf("Port not configured, use [%d]\n", DefaultPort))
		s.port = DefaultPort
	}
	return s
}

// Port can be bound by user
// user.Ports:
//
//	0 - all ports can be bind
//	80 - only 80 can be bind
//	80,22 - port 80 and 22 can be bind
//	1000-1010 - port from 1000 to 1020 can be bind
//
// port contained by user.Ports
func (s *ProxyServer) availabledPort(port int) bool {
	// TODO(shawnlu): Implement it
	return true
}

func (s *ProxyServer) getUserByUid(uid string) string {
	_, ok := s.users[uid]
	if ok {
		return uid
	}
	return ""
}

func (s *ProxyServer) getAuthedConn(authCtx string) connection.Connection {
	conn, ok := s.authedConn[authCtx]
	if ok {
		return conn
	}
	return nil
}

func (s *ProxyServer) auth(conn connection.Connection) (string, error) {
	// Parse pkt
	ctx := utils.NewTraceContext()
	cArrs := conn.GetArrs()
	pkt, err := protocol.ReadFromConn(cArrs.Conn)
	if err != nil {
		return "", fmt.Errorf("auth connection %s", err.Error())
	}

	// Switch req code
	switch pkt.GetPCode() {
	case protocol.RepAuth:
		uid := pkt.GetPayload().String()
		if len(s.getUserByUid(uid)) == 0 {
			return "", fmt.Errorf("no such user [%s]", uid)
		}

		// Generate auth ctx
		authCtx := uuid.NewV4().String()
		return authCtx, nil
	case protocol.RepPConn:
		authCtx := pkt.GetPayload().String()
		conn := s.getAuthedConn(authCtx)

		if conn != nil {
			logger.Error(ctx, "connection with auth ctx [%s] not exist, maybe staled")
			return "", nil
		}
		// Set connection NewPConn to true
		conn.SetToProxyConn()
		return "", nil
	}
	return "", nil
}

func (s *ProxyServer) bind(conn connection.Connection) (int, error) {
	return 8888, nil
}

func (s *ProxyServer) serveConn(conn connection.Connection) {
	defer func() {
		if r := recover(); r != nil {
			ctx := utils.NewTraceContext()
			cArrs := conn.GetArrs()

			logger.Error(ctx, fmt.Sprintf("server connection [%s] error", cArrs.Conn.RemoteAddr().String()))
			logger.Error(ctx, string(debug.Stack()))

			delete(s.authedConn, cArrs.AuthCtx)
			conn.Close()
		}
	}()

	// Auth
	authCtx, err := s.auth(conn)
	if err != nil {
		panic(err)
	}

	// For proxy connection, do io switch
	cArrs := conn.GetArrs()
	if cArrs.ProxyConn {
		aConn := s.getAuthedConn(authCtx)
		if aConn != nil {
			aConn.NewPConn(cArrs.Conn)
		}
		return
	}

	// For negotation connection bind then proxy
	bPort, err := s.bind(conn)
	if err != nil {
		panic(err)
	}
	err = conn.BindAndProxy(bPort)
	if err != nil {
		panic(err)
	}
}

func (s *ProxyServer) serve() error {
	defer s.ln.Close()

	for {
		ctx := utils.NewTraceContext()
		conn, err := s.ln.Accept()
		if err != nil {
			logger.Error(ctx, err.Error())
			continue
		}

		var c connection.Connection = connection.NewServerConnection(conn)
		go s.serveConn(c)
	}
}

// Launch server
func (s *ProxyServer) Launch() error {
	// Listen port and serve
	ctx := utils.NewTraceContext()
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Listen port %d %s", s.port, err.Error()))
		return err
	}
	s.ln = ln

	// Serve
	return s.serve()
}

func (s *ProxyServer) Stop() {
	s.ln.Close()
}
