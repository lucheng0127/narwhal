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

type clientError struct {
	msg string
}

func (err *clientError) Error() string {
	return fmt.Sprintf("Client error %s", err.msg)
}

type clientObj struct {
	remote struct {
		conn   net.Conn
		status string
	}
	local struct {
		conn   net.Conn
		status string
	}
	mux sync.Mutex
}

func waitRegistryReply(client *clientObj, targetPort, timeout int, wg *sync.WaitGroup) {
	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(timeout*int(time.Second)))
	defer cancel()

	done := make(chan bool)

	go func() {
		for {
			pkt, err := getPktFromConn(client.remote.conn)
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
				client.mux.Lock()
				client.remote.status = "FAILED"
				client.mux.Unlock()
				break
			}
			client.mux.Lock()
			client.remote.status = "ESTABLISHED"
			client.mux.Unlock()
			break
		}
		done <- true
	}()

	select {
	case <-ctx.Done():
		client.mux.Lock()
		client.remote.status = "FAILED"
		client.mux.Unlock()
		log.Warn("Waiting for registry reply timeout")
		break
	case <-done:
		client.mux.Lock()
		client.remote.status = "ESTABLISHED"
		client.mux.Unlock()
		log.Info("Registry client succeed")
		break
	}
	wg.Done()
}

func registryClient(client *clientObj, targetPort, maxRetryTimes, timeout int) {
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
	go waitRegistryReply(client, targetPort, timeout, &wg)
	wg.Add(1)

	// Keep registy client
	failedTimes := 0
	for {
		_, err = client.remote.conn.Write(pktBytes)
		if err != nil {
			failedTimes += 1
			log.Warnf("Failed to registry client %d times", failedTimes)
			if failedTimes >= maxRetryTimes {
				panic(fmt.Sprintf("Failed to registry client to server after %d times regtry, exit", failedTimes))
			}
		}
		client.mux.Lock()
		client.remote.status = "CONNTING"
		client.mux.Unlock()
		goto REGRETURN
	}
REGRETURN:
	wg.Wait()
}

func forwardTraffic(client *clientObj, forwardPort int) error {
	log.Info("Forward traffic between remote tcp connection and localport")
	// TODO:(lucheng) Implement it

	// Run goroutine listen remote and  get traffic from remote
	// send to localChannel, and get traffic from remoteChannel
	// send to remote

	// Run gorouting list local port and get traffic from local
	// send to remoteChannel, and get traffice from localChannel
	// send to local
	return nil
}

func RunClient(conf *internal.ClientConf) error {
	log.Infof("Launch client with config: %+v", *conf)
	client := new(clientObj)

	// Dial server
	serverAddr := fmt.Sprintf("%s:%d", conf.RemoteAddr, conf.ServerPort)
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		return &clientError{msg: err.Error()}
	}
	client.remote.conn = conn

REGISGTRY:
	// Keep registry client until succeed
	registryClient(client, conf.LocalPort, conf.MaxRetryTimes, conf.ReplyTimeout)
	if client.remote.status != "ESTABLISHED" {
		goto REGISGTRY
	}

	errGroup := new(errgroup.Group)
	// Launch a heartbeat job, running in background

	// Try to forward socket traffic
	errGroup.Go(func() error {
		err := forwardTraffic(client, conf.LocalPort)
		if err != nil {
			return &clientError{msg: err.Error()}
		}
		return nil
	})
	if err := errGroup.Wait(); err != nil {
		return err
	}
	return nil
}
