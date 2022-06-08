package service

import (
	"fmt"
	"narwhal/internal"
	"narwhal/proto"
	"net"

	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

var clientCm connManager

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
	}
}

func mointorConn(conn net.Conn) error {
	for {
		errGroup := new(errgroup.Group)
		for {
			errGroup.Go(func() error {
				err := handlePkt(conn, "client")
				if err != nil {
					return err
				}
				return nil
			})
			if err := errGroup.Wait(); err != nil {
				return err
			}
		}
	}
}

func RunClient(conf *internal.ClientConf) error {
	log.Infof("Launch client with config: %+v", *conf)
	clientCm.connMap = make(map[string]*connection)
	clientCm.lisMap = make(map[int]*lister)
	errGroup := new(errgroup.Group)

	// Dial server
	serverAddr := fmt.Sprintf("%s:%d", conf.RemoteAddr, conf.ServerPort)
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		return &clientError{msg: err.Error()}
	}
	clientCm.transferConnKey = fmt.Sprintf("forwardPort-%d", conf.LocalPort)
	transferConn := new(connection)
	transferConn.conn = conn
	transferConn.status = S_DUBIOUS
	clientCm.connMap[clientCm.transferConnKey] = transferConn

	// Monitor connection
	errGroup.Go(func() error {
		err := mointorConn(conn)
		if err != nil {
			return err
		}
		return nil
	})

	// TODO(lucheng): Launch a heartbeat job, running in background

REGISGTRY:
	// Keep registry client until succeed
	registryClient(conf.RemotePort, conf.MaxRetryTimes, conf.ReplyTimeout)
	if clientCm.connMap[conn.LocalAddr().String()].status != S_READY {
		goto REGISGTRY
	}

	// Check errors
	if err := errGroup.Wait(); err != nil {
		return &clientError{msg: err.Error()}
	}
	return nil
}
