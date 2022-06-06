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
			pkt, err := getPktFromConn(clientCm.connMap[targetPort].remote.conn)
			if err != nil {
				log.Debug("Failed to load narwhal packet from conn")
				continue
			}
			if pkt.Flag != proto.FLG_REP {
				continue
			}
			if pkt.TargetPort != uint16(targetPort) {
				continue
			}
			if pkt.Option != proto.OPT_OK {
				clientCm.mux.Lock()
				clientCm.connMap[targetPort].remote.status = CONN_UNHEALTH
				clientCm.mux.Unlock()
				break
			}
			clientCm.mux.Lock()
			clientCm.connMap[targetPort].remote.status = CONN_READY
			clientCm.mux.Unlock()
			break
		}
		done <- true
	}()

	select {
	case <-ctx.Done():
		clientCm.mux.Lock()
		clientCm.connMap[targetPort].remote.status = CONN_UNHEALTH
		clientCm.mux.Unlock()
		log.Warn("Waiting for registry reply timeout")
		break
	case <-done:
		clientCm.mux.Lock()
		clientCm.connMap[targetPort].remote.status = CONN_READY
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
	pkt.TargetPort = uint16(targetPort)
	pkt.Option = uint8(0)
	pkt.SetPayload([]byte("Registry client packet"))
	err := pkt.SetNoise()
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
		_, err = clientCm.connMap[targetPort].remote.conn.Write(pktBytes)
		if err != nil {
			failedTimes += 1
			log.Warnf("Failed to registry client %d times", failedTimes)
			if failedTimes >= maxRetryTimes {
				panic(fmt.Sprintf("Failed to registry client to server after %d times regtry, exit", failedTimes))
			}
		}
		clientCm.mux.Lock()
		clientCm.connMap[targetPort].remote.status = CONN_PENDING
		clientCm.mux.Unlock()
		goto REGRETURN
	}
REGRETURN:
	wg.Wait()
}

func RunClient(conf *internal.ClientConf) error {
	log.Infof("Launch client with config: %+v", *conf)
	clientCm.connMap = make(map[int]*connPeer)
	clientCm.connMap[conf.LocalPort] = new(connPeer)

	// Dial server
	serverAddr := fmt.Sprintf("%s:%d", conf.RemoteAddr, conf.ServerPort)
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		return &clientError{msg: err.Error()}
	}
	clientCm.connMap[conf.LocalPort].remote = &connection{
		conn:   conn,
		status: CONN_PENDING,
	}

REGISGTRY:
	// Keep registry client until succeed
	registryClient(conf.LocalPort, conf.MaxRetryTimes, conf.ReplyTimeout)
	if clientCm.connMap[conf.LocalPort].remote.status != CONN_READY {
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
