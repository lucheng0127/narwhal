package service

import (
	"errors"
	"fmt"
	"narwhal/internal"
	"narwhal/proto"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

func RunHeartBeatJob(conf *internal.ClientConf) error {
	failedTimes := 0
	ticker := time.NewTicker(time.Duration(conf.HeartBeatInterval * int(time.Second)))
	for range ticker.C {
		serverAddr := fmt.Sprintf("%s:%d", conf.RemoteAddr, conf.ServerPort)
		err := proto.SendHeartBeat(serverAddr, conf.LocalPort)
		if err != nil {
			log.Warnf("Send heartbeat failed %s", err)
			failedTimes += 1
		} else {
			log.Debug("Sent heartbeat")
			failedTimes = 0
		}
		if failedTimes >= 5 {
			break
		}
	}
	return errors.New("max retry times for failed to send heartbeat")
}

func RunClient(conf *internal.ClientConf) error {
	log.Infof("Launch client with config: %+v", *conf)
	eGroup := new(errgroup.Group)
	eGroup.Go(func() error {
		return RunHeartBeatJob(conf)
	})
	if err := eGroup.Wait(); err != nil {
		return err
	}
	return nil
}
