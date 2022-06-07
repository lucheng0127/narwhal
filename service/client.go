package service

import (
	"context"
	"fmt"
	"narwhal/internal"
	"narwhal/proto"
	"net"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

var clientCm connManager

func waitRegistryReply(targetPort, timeout int, wg *sync.WaitGroup) {
	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(timeout*int(time.Second)))
	defer cancel()

	done := make(chan bool)

	go func() {
		for {
			pkt, err := getPktFromConn(clientCm.connMap[clientCm.transferConnKey].conn)
			if err != nil {
				log.Debug("Failed to load narwhal packet from conn")
				continue
			}
			if pkt.Flag != proto.FLG_REP {
				continue
			}
			if pkt.Code != proto.C_OK {
				// Registy failed return
				break
			}
			clientCm.mux.Lock()
			clientCm.connMap[clientCm.transferConnKey].status = S_READY
			clientCm.mux.Unlock()
			break
		}
		done <- true
	}()

	select {
	case <-ctx.Done():
		log.Warn("Waiting for registry reply timeout")
		break
	case <-done:
		clientCm.mux.Lock()
		clientCm.connMap[clientCm.transferConnKey].status = S_READY
		clientCm.mux.Unlock()
		log.Info("Registry client succeed")
		break
	}
	wg.Done()
}

func registryClient(targetPort, maxRetryTimes, timeout int) {
	// Build registry packet
	pkt := new(proto.NWPacket)
	pkt.Flag = proto.FLG_REG
	pkt.SetUnassignedAddrs()
	pkt.Code = proto.C_OK
	err := pkt.SetTargetPort(int16(targetPort))
	if err != nil {
		panic(fmt.Sprintf("Set target port to payload error %s", err))
	}
	err = pkt.SetNoise()
	if err != nil {
		panic(fmt.Sprintf("Registry client error %s", err))
	}
	pktBytes, err := pkt.Encode()
	if err != nil {
		panic(fmt.Sprintf("Registry client error %s", err))
	}

	var wg sync.WaitGroup
	go waitRegistryReply(targetPort, timeout, &wg)
	wg.Add(1)

	// Keep registy client
	failedTimes := 0
	for {
		_, err = clientCm.connMap[clientCm.transferConnKey].conn.Write(pktBytes)
		if err != nil {
			failedTimes += 1
			log.Warnf("Failed to registry client %d times", failedTimes)
			if failedTimes >= maxRetryTimes {
				panic(fmt.Sprintf("Failed to registry client to server after %d times regtry, exit", failedTimes))
			}
		}
		goto REGRETURN
	}
REGRETURN:
	wg.Wait()
}

func RunClient(conf *internal.ClientConf) error {
	log.Infof("Launch client with config: %+v", *conf)
	clientCm.connMap = make(map[string]*connection)
	clientCm.lisMap = make(map[int]*lister)

	// Dial server
	serverAddr := fmt.Sprintf("%s:%d", conf.RemoteAddr, conf.ServerPort)
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		return &clientError{msg: err.Error()}
	}
	clientCm.transferConnKey = conn.LocalAddr().String()
	transferConn := new(connection)
	transferConn.conn = conn
	transferConn.status = S_DUBIOUS
	clientCm.connMap[clientCm.transferConnKey] = transferConn

REGISGTRY:
	// Keep registry client until succeed
	registryClient(conf.LocalPort, conf.MaxRetryTimes, conf.ReplyTimeout)
	if clientCm.connMap[conn.LocalAddr().String()].status != S_READY {
		goto REGISGTRY
	}

	errGroup := new(errgroup.Group)
	// TODO(lucheng): Launch a heartbeat job, running in background

	// Try to forward socket traffic
	errGroup.Go(func() error {
		err := forwardTrafficClient()
		if err != nil {
			return err
		}
		return nil
	})

	// Check errors
	if err := errGroup.Wait(); err != nil {
		return &clientError{msg: err.Error()}
	}
	return nil
}
