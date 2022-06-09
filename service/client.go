package service

import (
	"narwhal/internal"

	log "github.com/sirupsen/logrus"
)

type client struct {
	localPort int
}

var clientObj client

func RunClient(conf *internal.ClientConf) error {
	log.Infof("Launch client with config: %+v", *conf)
	clientObj.localPort = conf.LocalPort
	// Registry client, panic error
	// Monitor connection forever, handle pkt
	return nil
}
