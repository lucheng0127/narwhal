package proxy

import (
	"fmt"
	"net"
	"os"
	"runtime/debug"

	logger "github.com/lucheng0127/narwhal/internal/pkg/log"
	"github.com/lucheng0127/narwhal/internal/pkg/utils"
	"github.com/lucheng0127/narwhal/pkg/connection"
	"github.com/lucheng0127/narwhal/pkg/protocol"
	uuid "github.com/satori/go.uuid"
)

type ProxyServer struct {
	port       int // Service port
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
		logger.Error(ctx, "Port not configured")
		return nil
	}
	return s
}

func (s *ProxyServer) Stop() {
	os.Exit(0)
}

func (s *ProxyServer) ServeConn(conn connection.Connection) {
	defer func() {
		if r := recover(); r != nil {
			ctx := utils.NewTraceContext()
			logger.Error(ctx, string(debug.Stack()))
			conn.Close()
		}
	}()

	ctx := utils.NewTraceContext()
	go conn.Serve(ctx)

	// Wait client send auth and bind cmd, if proxy connection,
	// will send CmdNewConn cmd, after handle it will also send true to ready
	<-conn.(*connection.SConn).Ready

	if !conn.ShouldProxy() {
		// There are two kind of connection, authed connection and proxy connection
		// authed connection: is the first connection that server and client establish,
		// use it auth, bind port
		// proxy connection: when a new connection establish to server binding port,
		// server will notify client this a new connection, client will establish a
		// new connection to local target port, after that client will make a new connection
		// with server, when connection establish send server a CmdNewConn, then start proxy
		return
	}

	// Auth
	if !s.Auth(conn.GetUID()) {
		logger.Error(ctx, "client auth failed after 5 times retry, close connection")
		err := conn.ReplayWithCode(ctx, protocol.RepAuthFailed)
		panic(err)
	}
	// Generate auth ctx
	authctx := uuid.NewV4().String()
	conn.SetAuthCtx(authctx)
	s.authedConn[authctx] = conn
	conn.ReplayWithAuthCtx(ctx, authctx)

	// Check port validate
	if !s.AvailabledPort(conn.GetBindPort()) {
		logger.Error(ctx, "not premmited binding port")
		err := conn.ReplayWithCode(ctx, protocol.RepInvalidPort)
		panic(err)
	}
	// Bind port and proxy
	go conn.Proxy()
}

// Get proxy connection from GConnPool then send it to connection.ProxyConnCh
//
// proxy connection:
//
//	when a new connection establish will send server cmd CmdNewConn,
//	with authCtx as payload, server use payload get authed connection
//	and send new connection to authed connection.ProxyConnCh.
//	authed connection will do proxy with this new connection
func (s *ProxyServer) Monitor() {
	ctx := utils.NewTraceContext()
	for {
		authCtx := <-connection.GConnPool.Pool
		proxyConn, ok := connection.GConnPool.PoolMap[authCtx]
		if !ok {
			logger.Warn(ctx, fmt.Sprintf("no such connection in GConnPool with key [%s]\n", authCtx))
			continue
		}

		authedConn, ok := s.authedConn[authCtx]
		if !ok {
			logger.Warn(ctx, fmt.Sprintf("no auth connection with contex [%s]\n", authCtx))
			continue
		}

		authedConn.(*connection.SConn).ProxyConnCh <- proxyConn
	}
}

func (s *ProxyServer) Serve(ln net.Listener) error {
	defer ln.Close()

	for {
		ctx := utils.NewTraceContext()
		conn, err := ln.Accept()
		if err != nil {
			logger.Error(ctx, err.Error())
			continue
		}

		var c connection.Connection = connection.NewServerConnection(conn)
		go s.ServeConn(c)
	}
}

func (s *ProxyServer) Launch() error {
	// Listen port and serve
	ctx := utils.NewTraceContext()
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Listen port %d %s", s.port, err.Error()))
		return err
	}

	return s.Serve(ln)
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
func (s *ProxyServer) AvailabledPort(port int) bool {
	// TODO(shawnlu): Implement it
	return true
}

func (s *ProxyServer) Auth(uID string) bool {
	if _, ok := s.users[uID]; ok {
		return true
	}
	return false
}
