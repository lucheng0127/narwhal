package service

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"narwhal/internal"
	"narwhal/proto"
	"net"

	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

func registryTargetPort(conn *Connection, targetPort int) error {
	// Send registry packet
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, uint16(targetPort))
	if err != nil {
		return internal.NewError("Registry client", err.Error())
	}

	// New narwhal packet
	pkt := new(proto.NWPacket)
	pkt.Flag = proto.FLG_REG
	pkt.Seq = conn.Key // For registry pkt Seq key have no means
	pkt.Result = proto.RST_OK
	pkt.SetPayload(buf.Bytes())
	err = pkt.SetNoise()
	if err != nil {
		return internal.NewError("Registry client", err.Error())
	}

	pktBytes, err := pkt.Encode()
	if err != nil {
		return internal.NewError("Registry client", err.Error())
	}

	// Send reg pkt, if registry failed panic in handle reply
	_, err = conn.Conn.Write(pktBytes)
	if err != nil {
		return internal.NewError("Registry client", err.Error())
	}
	log.Debugf("Send registry pkt for target port %d", targetPort)
	return nil
}

func handleForwardConn(conn *Connection) {
	for {
		if conn.Status == C_CLOSED {
			panic("Connection to forward port closed, please make sure forward port is listen")
		}
		// Fetch data to packet
		pktBytes, err := fetchDataToPktBytes(conn)
		if err != nil {
			panic(err)
		}

		// Send to transfer connection
		transferConn, ok := CM.TConnMap[CM.ServerAddr]
		if !ok {
			panic("Connection to narwhal server broken")
		}

		_, err = transferConn.Write(pktBytes)
		if err != nil {
			log.Errorf("Send packet bytes to transfer connection error\n%s", err.Error())
			continue
		}
	}
}

func handleClientConn(conn *Connection) error {
	for {
		pkt, err := fetchPkt(conn)
		if err != nil {
			return err
		}
		if pkt == nil {
			// Connection closed out loop
			break
		}

		switch pkt.Flag {
		case proto.FLG_DAT:
			handleDataClient(pkt)
		case proto.FLG_REP:
			handleReply(pkt)
		}
	}
	return nil
}

func RunClient(conf *internal.ClientConf) error {
	// Connection to narwhal
	serverAddr := fmt.Sprintf("%s:%d", conf.RemoteAddr, conf.ServerPort)
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		return internal.NewError("Connect to narwhal server error", err.Error())
	}
	CM.ClientLocalPort = conf.LocalPort
	CM.ServerAddr = conn.RemoteAddr().String()
	newConn := new(Connection)
	newConn.Conn = conn
	newConn.Key = newSeq()
	// No need send conn to ConnMap

	// Set TConnMap, serverAddr as key
	CM.Mux.Lock()
	CM.TConnMap[conn.RemoteAddr().String()] = conn
	CM.Mux.Unlock()
	errGup := new(errgroup.Group)
	log.Infof("New connection to %s local %s",
		newConn.Conn.RemoteAddr().String(), newConn.Conn.LocalAddr().String())

	// Groutine: Monitor conn and handle pkt
	errGup.Go(func() error {
		err := handleClientConn(newConn)
		if err != nil {
			return err
		}
		return nil
	})

	// Registry client with target port
	err = registryTargetPort(newConn, conf.RemotePort)
	if err != nil {
		return err
	}

	if err := errGup.Wait(); err != nil {
		return err
	}
	return nil
}
