package service

import (
	"errors"
	"narwhal/internal"
	"narwhal/proto"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type Client struct {
	ServerPort    int // Port that server listen
	RemoteAddr    string
	RemotePort    int // Port that server forward traffic from it to LocalPort
	LocalAddr     string
	LocalPort     int
	Interval      int // HeartBeat send interval
	MaxRetryTimes int
	Status        string
	CTXTimeout    int // Context timeout
}

func RunHeartBeatJob(client *Client) error {
	failedTimes := 0
	ticker := time.NewTicker(time.Duration(client.Interval * int(time.Second)))
	for range ticker.C {
		err := proto.SendHeartBeat(client.RemoteAddr, client.ServerPort)
		if err != nil {
			log.Warnf("Send heartbeat failed %s", err)
			failedTimes += 1
		} else {
			log.Debug("Sent heartbeat")
			failedTimes = 0
		}
		if failedTimes >= 5 {
			client.Status = "DEAD"
			break
		}
	}
	return errors.New("max retry times for failed to send heartbeat")
}

func RunClient(conf *internal.ClientConf) error {
	log.Infof("Launch client with config: %+v", *conf)
	client := Client{
		ServerPort:    conf.ServerPort,
		RemoteAddr:    conf.RemoteAddr,
		RemotePort:    conf.RemotePort,
		LocalAddr:     conf.LocalAddr,
		LocalPort:     conf.LocalPort,
		Interval:      conf.HeartBeatInterval,
		MaxRetryTimes: conf.MaxRetryTimes,
		Status:        "HEALTHY",
		CTXTimeout:    conf.CTXTimeOut,
	}

	eGroup := new(errgroup.Group)
	eGroup.Go(func() error {
		return RunHeartBeatJob(&client)
	})
	if err := eGroup.Wait(); err != nil {
		return err
	}
	return nil
}
