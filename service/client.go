package service

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"narwhal/internal"
	"narwhal/proto"
	"net"
	"sync"

	log "github.com/sirupsen/logrus"
)

func registryClient(conn net.Conn, targetPort int) error {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, uint16(targetPort))
	if err != nil {
		return internal.NewError("Registry client", err.Error())
	}

	pkt, err := newNWPkt(proto.FLG_REG, newSeq(), buf.Bytes())
	if err != nil {
		return internal.NewError("Registry client", err.Error())
	}
	pktBytes, err := pkt.Encode()
	if err != nil {
		return internal.NewError("Registry client", err.Error())
	}

	n, err := conn.Write(pktBytes)
	if err != nil {
		return internal.NewError("Registry client", err.Error())
	}
	log.Debugf("Send registy packet %d bytes", n)
	return nil
}

func RunClient(conf *internal.ClientConf) error {
	log.Infof("Launch client with config: %+v", *conf)
	CM.ClientLocalPort = conf.LocalPort
	// Registry client, panic error
	serverAddr := fmt.Sprintf("%s:%d", conf.RemoteAddr, conf.ServerPort)
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		return internal.NewError("Connect to narwhal server error", err.Error())
	}
	CM.Mux.Lock()
	newConn := new(Connection)
	newConn.Conn = conn
	newConn.Key = newSeq()
	CM.ConnMap[newConn.Key] = newConn
	CM.TransferConnMap[CM.ClientLocalPort] = conn
	CM.Mux.Unlock()

	// Monitor connection forever, handle pkt, run before send reg pkt
	var wg sync.WaitGroup
	go monitorConn(newConn, int(MOD_T_C))
	wg.Add(1)

	err = registryClient(conn, conf.RemotePort)
	if err != nil {
		return err
	}

	wg.Wait()
	return nil
}
