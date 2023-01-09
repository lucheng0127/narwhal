package server

import (
	"context"
	"fmt"

	logger "github.com/lucheng0127/narwhal/internal/pkg/log"
)

type options struct {
	port int
}

type ServerOption func(*options)

func ListenPort(port int) ServerOption {
	return func(o *options) {
		o.port = port
	}
}

type Server struct {
	port int
}

func NewServer(opt ...ServerOption) *Server {
	server := new(Server)
	opts := new(options)
	for _, o := range opt {
		o(opts)
	}

	ctx := context.Background()
	if opts.port == 0 {
		logger.Error(ctx, "Port not configured")
		return nil
	}
	server.port = opts.port
	return server
}

func (s *Server) Launch() {
	fmt.Println("Launched")
}

func (s *Server) Stop() {
}
