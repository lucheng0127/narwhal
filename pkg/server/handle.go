package server

import (
	"fmt"
	"net"

	logger "github.com/lucheng0127/narwhal/internal/pkg/log"
	"github.com/lucheng0127/narwhal/internal/pkg/utils"
)

func (s *Server) HandleConn(conn net.Conn) {
	defer conn.Close()
	ctx := utils.NewTraceContext()
	logger.Info(ctx, fmt.Sprintf("New connection from %s", conn.RemoteAddr().String()))

	// Parse request

	// Auth

	// Start proxy
}
