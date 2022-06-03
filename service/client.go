package service

import (
	"fmt"
	"narwhal/internal"
	"narwhal/proto"
	"net"

	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type clientError struct {
	msg string
}

func (err *clientError) Error() string {
	return fmt.Sprintf("Client error %s", err.msg)
}

func registryClient(conn net.Conn, targetPort, maxRetryTimes, timeout int) error {
	// Build registry packet
	pkt := new(proto.NWPacket)
	pkt.Flag = proto.FLG_REG
	pkt.TargetPort = uint16(targetPort)
	pkt.Option = uint8(0)
	pkt.SetPayload([]byte("Registry client packet"))
	err := pkt.SetNoise()
	if err != nil {
		log.Errorf("Set noise for narwhal packet error $s", err)
		return err
	}
	pktBytes, err := pkt.Encode()
	if err != nil {
		log.Error("Failed to encode narwhal packet")
		return err
	}

	// TODO(lucheng): run goroutine wiating for reply

	// Keep registy client
	// Use errRaise determine raise err or nil when goto RETURN
	errRaised := false
	failedTimes := 0
	for {
		_, err = conn.Write(pktBytes)
		if err != nil {
			failedTimes += 1
			log.Warnf("Failed to registry client %d times", failedTimes)
			if failedTimes >= maxRetryTimes {
				errRaised = true
				// Send registry failed goto return and raise error
				goto RETURN
			}
		}
		goto RETURN
	}

RETURN:
	if errRaised {
		return fmt.Errorf("failed to registry client to server after %d times retry", maxRetryTimes)
	}
	return nil
}

func forwardTraffic(conn net.Conn, forwardPort int) error {
	return nil
}

func RunClient(conf *internal.ClientConf) error {
	log.Infof("Launch client with config: %+v", *conf)
	// Dial server
	serverAddr := fmt.Sprintf("%s:%d", conf.RemoteAddr, conf.ServerPort)
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		return &clientError{msg: err.Error()}
	}

	// Keep registry client until succeed
	err = registryClient(conn, conf.LocalPort, conf.MaxRetryTimes, conf.CTXTimeOut)
	if err != nil {
		return &clientError{msg: err.Error()}
	}
	log.Info("Registry client succeed")

	errGroup := new(errgroup.Group)
	// Launch a heartbeat job, running in background

	// Try to forward socket traffic
	errGroup.Go(func() error {
		err := forwardTraffic(conn, conf.LocalPort)
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
