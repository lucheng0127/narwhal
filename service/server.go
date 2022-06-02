package service

import (
	"fmt"
	"narwhal/internal"
	"narwhal/proto"
	"net"

	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

var serverHandles proto.HandleMap

func HandleConn(conn net.Conn, mtu int) {
	buf := make([]byte, mtu-20)
	n, err := conn.Read(buf)
	if err != nil {
		log.Errorf("Failed to read data from tcp connection %s", err)
	}

	log.Debugf("Read %d bytes from tcp conn:\nRemote info: %+v Local info: %+v", n, conn.RemoteAddr(), conn.LocalAddr())
	pkg := new(proto.NWPackage)
	err = pkg.Unmarshal(buf[:n])
	if err != nil {
		log.Errorf("Failed to parse narwhal package %s", err)
		return
	}
	err = serverHandles[pkg.Flag](conn, pkg)
	if err != nil {
		log.Errorf("Handle pkg %s", err)
		return
	}
}

func ListenServer(port, mtu int) error {
	listen, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Panicf("Failed to setup tcp listen server %s", err)
		return err
	}
	// Registry package handles for server
	serverHandles = proto.GetHandles("server")
	log.Infof("Start to listen port: %d", port)
	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Warn("Failed to build connection %s", err)
			continue
		}

		go HandleConn(conn, mtu)
	}
}

func RunServer(conf *internal.ServerConf) error {
	eGroup := new(errgroup.Group)

	eGroup.Go(func() error {
		return ListenServer(conf.ListenPort, conf.MTU)
	})
	if err := eGroup.Wait(); err != nil {
		return err
	}
	return nil
}
