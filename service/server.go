package service

import (
	"fmt"
	"narwhal/internal"
	"net"

	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

var serverCm connManager

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
			err = handlePkt(conn, "server")
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

	if err := errGroup.Wait(); err != nil {
		return err
	}
	return nil
}
