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

type Server interface {
	Launch() error
	Serve(net.Listener) error
	ServeConn(conn connection.Connection)
}

type ProxyServer struct {
	port       int // Service port
	users      map[string]string
	authedConn map[string]connection.Connection
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
	go conn.Serve()

	// Wait client send auth and bind cmd
	<-conn.(*connection.SConn).Ready

	// Auth
	if !s.Auth(conn.(*connection.SConn).UID) {
		logger.Error(ctx, "client auth failed after 5 times retry, close connection")
		conn.ReplayWithCode(protocol.RepAuthFailed)
		conn.Close()
	}
	// Generate auth ctx
	authctx := uuid.NewV4().String()
	conn.(*connection.SConn).AuthCtx = authctx
	s.authedConn[authctx] = conn
	conn.ReplayWithAuthCtx()

	// Check port validate
	if !s.AvailabledPort(conn.(*connection.SConn).BindPort) {
		logger.Error(ctx, "not premmited binding port")
		conn.ReplayWithCode(protocol.RepInvalidPort)
		conn.Close()
	}
	// Bind port and proxy
	go conn.Proxy()
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
