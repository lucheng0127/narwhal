package service

import (
	"narwhal/internal"

	log "github.com/sirupsen/logrus"
)

func RunServer(conf *internal.ServerConf) error {
	log.Infof("Launch server with config: %+v", *conf)
	return nil
}
