package proxy

import (
	"fmt"
	"net"
	"runtime/debug"
	"strconv"
	"strings"

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
func (s *ProxyServer) availabledPort(uid string, port int) bool {
	pr, ok := s.users[uid]
	if !ok {
		return false
	}

	if strings.Contains(pr, "-") {
		prArray := strings.Split(pr, "-")
		if len(prArray) != 2 {
			return false
		}

		prL, err := strconv.Atoi(prArray[0])
		if err != nil {
			return false
		}
		prR, err := strconv.Atoi(prArray[1])
		if err != nil {
			return false
		}

		if prL <= port && port <= prR {
			return true
		}
		return false
	}

	if strings.Contains(pr, ",") {
		prArray := strings.Split(pr, ",")
		for _, prIStr := range prArray {
			prI, err := strconv.Atoi(prIStr)
			if err == nil && prI == port {
				return true
			}
		}
		return false
	}

	prI, err := strconv.Atoi(pr)
	if err == nil {
		if prI == 0 || prI == port {
			return true
		}
	}

	return false
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
	cArrs := conn.GetArrs()
	pkt, err := protocol.ReadFromConn(cArrs.Conn)
	if err != nil {
		return "", fmt.Errorf("parse auth request %s", err.Error())
	}

	// Switch req code
	switch pkt.GetPCode() {
	case protocol.RepAuth:
		uid := pkt.GetPayload().String()
		if len(s.getUserByUid(uid)) == 0 {
			return "", fmt.Errorf("no such user [%s]", uid)
		}
		conn.SetUID(uid)

		// Generate auth ctx
		authCtx := uuid.NewV4().String()
		conn.SetAuthCtx(authCtx)

		// TODO(Reply)
		return authCtx, nil
	case protocol.RepPConn:
		authCtx := pkt.GetPayload().String()
		conn := s.getAuthedConn(authCtx)

		if conn == nil {
			return "", fmt.Errorf("connection with auth ctx [%s] not exist, maybe staled", authCtx)
		}
		// Set connection NewPConn to true
		conn.SetToProxyConn()

		// TODO(Reply)
		return "", nil
	default:
		return "", fmt.Errorf("invalidate auth request format")
	}
}

func (s *ProxyServer) bind(conn connection.Connection) (int, error) {
	cArrs := conn.GetArrs()
	pkt, err := protocol.ReadFromConn(cArrs.Conn)
	if err != nil {
		return -1, fmt.Errorf("parse bind request %s", err.Error())
	}

	switch pkt.GetPCode() {
	case protocol.RepBind:
		// TODO(Reply)
		bPort := pkt.GetPayload().Int()
		if bPort == -1 {
			return -1, fmt.Errorf("invalidate bind request, binding port not set")
		}

		if !s.availabledPort(conn.GetArrs().UID, bPort) {
			return -1, fmt.Errorf("not permitted binding port [%d]", bPort)
		}
		return bPort, nil
	default:
		return -1, fmt.Errorf("invalidate binding request format")
	}
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
			return
		}
	}()

	ctx := utils.NewTraceContext()
	// Auth
	authCtx, err := s.auth(conn)
	if err != nil {
		logger.Error(ctx, err.Error())
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
		logger.Error(ctx, err.Error())
		panic(err)
	}
	err = conn.BindAndProxy(bPort)
	if err != nil {
		logger.Error(ctx, err.Error())
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
