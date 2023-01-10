package server

import (
	"context"
	"fmt"
	"net"

	logger "github.com/lucheng0127/narwhal/internal/pkg/log"
	"github.com/lucheng0127/narwhal/internal/pkg/utils"
)

type Server struct {
	port int
}

func NewServer(opt ...ServerOption) *Server {
	server := new(Server)
	for _, o := range opt {
		o(server)
	}

	ctx := context.Background()
	if server.port == 0 {
		logger.Error(ctx, "Port not configured")
		return nil
	}
	return server
}

func (s *Server) Serve(l net.Listener) error {
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}

		// Handle connection
		go s.HandleConn(conn)
	}
}

func (s *Server) Launch() {
	ctx := utils.NewTraceContext()

	// Listen port
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		logger.Error(ctx, err.Error())
		s.Stop()
	}

	// Run socks5 server
	if err = s.Serve(l); err != nil {
		logger.Painc(ctx, err.Error())
	}
}

func (s *Server) Stop() {
	// TODO(shawnlu): Implement it
}
