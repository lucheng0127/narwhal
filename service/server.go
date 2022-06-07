package service

import (
	"fmt"
	"narwhal/internal"
	"net"

	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

var serverCm connManager

func handleConn(conn net.Conn) error {
	// Get server handles map
	handle := handleManager("server")

	errGroup := new(errgroup.Group)

	// Fetch narwhal packet
	pkt, err := getPktFromConn(conn)
	if err != nil {
		return err
	}

	// Handle packet goroutine
	errGroup.Go(func() error {
		err := handle[pkt.Flag](conn, pkt)
		if err != nil {
			return err
		}
		return nil
	})
	return nil
}

func LaunchTCPServer(port int) error {
	// Listen server
	listen, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return &serverError{msg: err.Error()}
	}
	log.Infof("Launch server on port: %d", port)

	// Run packet handle with goroutine
	errGroup := new(errgroup.Group)
	for {
		conn, err := listen.Accept()
		if err != nil {
			return err
		}

		errGroup.Go(func() error {
			err = handleConn(conn)
			if err != nil {
				return err
			}
			return nil
		})
		if err := errGroup.Wait(); err != nil {
			return err
		}
	}
}

func RunServer(conf *internal.ServerConf) error {
	log.Infof("Launch server with config: %+v", *conf)
	serverCm.connMap = make(map[string]*connection)
	serverCm.lisMap = make(map[int]*lister)

	errGroup := new(errgroup.Group)
	errGroup.Go(func() error {
		err := LaunchTCPServer(conf.ListenPort)
		if err != nil {
			return &tcpServerError{msg: err.Error()}
		}
		return nil
	})

	// TODO(lucheng): Add func forward traffic between local lister and
	// remote connections, data from serverCm.connMap, maybe need add a
	// tigger to active recheck serverCm.connMap
	// Health check set to unhealth close connection

	if err := errGroup.Wait(); err != nil {
		return err
	}
	return nil
}
