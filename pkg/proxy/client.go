package proxy

import (
	"fmt"
	"net"

	logger "github.com/lucheng0127/narwhal/internal/pkg/log"
	"github.com/lucheng0127/narwhal/internal/pkg/utils"
	"github.com/lucheng0127/narwhal/pkg/connection"
)

type ClientServer struct {
	host   string
	rPort  uint16
	lPort  uint16
	uid    string
	client connection.Client
}

func NewClientServer(opts ...COption) Server {
	s := new(ClientServer)

	for _, o := range opts {
		o(s)
	}

	return s
}

func (c *ClientServer) Launch() error {
	// Connect to host
	ctx := utils.NewTraceContext()
	conn, err := net.Dial("tcp", c.host)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("connection to server [%s] %s", c.host, err.Error()))
		return err
	}
	c.client = connection.NewClient(conn)

	// Auth
	err = c.client.Auth(c.uid)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("auth %s", err.Error()))
		return err
	}

	// Bind port
	err = c.client.Bind(c.rPort)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("bind %s", err.Error()))
		return err
	}

	// Monitor and proxy
	return c.client.MonitorAndProxy(c.lPort)
}

func (c *ClientServer) Stop() {
	ctx := utils.NewTraceContext()
	logger.Info(ctx, "stop client server ...")
	c.client.Close()
}
