package service

import (
	"narwhal/internal"

	log "github.com/sirupsen/logrus"
)

type server struct {
	serverPort int
}

var serverObj server

func RunServer(conf *internal.ServerConf) error {
	log.Infof("Launch server with config: %+v", *conf)
	serverObj.serverPort = conf.ListenPort

	// Launch tcp server and listen forever
	// in it handle new connection
	err := launchNWServer(serverObj.serverPort)
	if err != nil {
		return err
	}
	return nil
}
