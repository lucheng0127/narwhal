package service

import (
	"narwhal/internal"
	"narwhal/rpc"
	"sync"

	log "github.com/sirupsen/logrus"
)

func RunServer(conf *internal.ServerConf) error {
	log.Infof("Launch server with config: %+v", *conf)
	rpcServer := rpc.RPCServer{}
	var wg sync.WaitGroup

	go rpc.LaunchRPCServer(&rpcServer, conf.ListenPort, wg)
	wg.Add(1)
	wg.Wait()
	return nil
}
