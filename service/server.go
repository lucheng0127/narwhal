package service

import (
	"fmt"
	"narwhal/internal"
	"net"

	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

// Registry port connection map

type serverError struct {
	msg string
}

type connHandleError struct {
	serverError
}

func (err *serverError) Error() string {
	return fmt.Sprintf("Narwhal server error %s", err.msg)
}

func (err *connHandleError) Error() string {
	return fmt.Sprintf("Handle TCP connection error %s", err.msg)
}

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
			log.Errorf("Accept tcp connection error %s", err)
			return &serverError{msg: err.Error()}
		}

		errGroup.Go(func() error {
			err = handleConn(conn)
			if err != nil {
				return &connHandleError{serverError{msg: err.Error()}}
			}
			return nil
		})
		if err := errGroup.Wait(); err != nil {
			return err
		}
	}
}

func RunServer(conf *internal.ServerConf) error {
	errGroup := new(errgroup.Group)
	errGroup.Go(func() error {
		err := LaunchTCPServer(conf.ListenPort)
		if err != nil {
			log.Errorf("Launch tcp server failed %s", err)
			return err
		}
		return nil
	})

	if err := errGroup.Wait(); err != nil {
		return err
	}
	return nil
}
