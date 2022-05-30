package service

import (
	"narwhal/internal"

	log "github.com/sirupsen/logrus"
)

func RunClient(conf *internal.ClientConf) error {
	log.Infof("Launch client with config: %+v", *conf)
	return nil
}
