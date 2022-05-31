package service

import (
	"fmt"
	"narwhal/internal"
	"narwhal/rpc"
	"time"

	log "github.com/sirupsen/logrus"
)

type Client struct {
	ServerAddr    string
	ServerPort    int
	TargetPort    int
	Interval      int
	MaxRetryTimes int
	Status        string
}

func newHeartBeatJob(client *Client, done chan bool) {
	failedTimes := 0
	ticker := time.NewTicker(time.Duration(client.Interval * int(time.Second)))
	for range ticker.C {
		log.Debugf("Send heartbeat for target port: %d", client.TargetPort)
		serverAddr := fmt.Sprintf("%s:%d", client.ServerAddr, client.ServerPort)
		ret := rpc.SendHeartBeat(serverAddr, client.TargetPort)
		if ret != 0 {
			failedTimes += 1
		} else {
			failedTimes = 0
		}
		if failedTimes >= 5 {
			client.Status = "UNHEALTH"
			break
		}
	}
	ticker.Stop()
	done <- true
}

func RunClient(conf *internal.ClientConf) error {
	log.Infof("Launch client with config: %+v", *conf)
	client := Client{
		ServerAddr:    conf.RemoteAddr,
		ServerPort:    conf.ServerPort,
		TargetPort:    conf.RemotePort,
		Interval:      conf.HeartBeatInterval,
		MaxRetryTimes: conf.MaxRetryTimes,
		Status:        "ACTIVE",
	}

	hbJobCh := make(chan bool)
	go newHeartBeatJob(&client, hbJobCh)

	hbJobDone := <-hbJobCh
	if hbJobDone {
		log.Warn("Narwhal heartbeat job stopped")
	}
	return nil
}
