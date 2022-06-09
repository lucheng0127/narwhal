package service

import (
	"encoding/binary"
	"fmt"
	"narwhal/internal"
	"narwhal/proto"
	"net"

	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type Callback func(conn net.Conn, pkt *proto.NWPacket) error

// Handlers for each narwhal packet

func handleRegistry(conn net.Conn, pkt *proto.NWPacket) error {
	// Get target pot registry it
	targetPort := int(binary.BigEndian.Uint16(pkt.Payload))
	CM.Mux.Lock()
	newConn := new(Connection)
	newConn.Conn = conn
	newConn.Key = pkt.Seq
	CM.ConnMap[pkt.Seq] = newConn
	CM.Mux.Unlock()

	errGup := new(errgroup.Group)
	errGup.Go(func() error {
		err := listenAndService(targetPort)
		if err != nil {
			return internal.NewError("Handle registry", err.Error())
		}
		return nil
	})

	// Reply
	repPkt, err := newNWPkt(proto.FLG_REP, pkt.Seq, pkt.Payload)
	if err != nil {
		return internal.NewError("Handle registry", err.Error())
	}
	pktBytes, err := repPkt.Encode()
	if err != nil {
		return internal.NewError("Handle registry", err.Error())
	}

	_, err = conn.Write(pktBytes)
	if err != nil {
		return internal.NewError("Handle registry", err.Error())
	}

	log.Debugf("Registry port %d succeed", targetPort)

	// Check err
	if err = errGup.Wait(); err != nil {
		return err
	}
	return nil
}

func handleReply(conn net.Conn, pkt *proto.NWPacket) error {
	if pkt.Result == proto.PORT_INUSE {
		panic("Target port used by others")
	}
	if pkt.Result != proto.RST_OK {
		panic("Registry failed")
	}
	log.Info("Registry succeed")
	return nil
}

func handleHeartBeat(conn net.Conn, pkt *proto.NWPacket) error {
	return nil
}

func handleDataServer(conn net.Conn, pkt *proto.NWPacket) error {
	log.Info(pkt.Seq)
	return nil
}

func handleDataClient(transferConn net.Conn, pkt *proto.NWPacket) error {
	errGup := new(errgroup.Group)

	// Get seq num
	_, ok := CM.ConnMap[pkt.Seq]
	if !ok {
		// Dial to local port, add into ConnMap
		conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", clientObj.localPort))
		if err != nil {
			return internal.NewError("Dial to local", err.Error())
		}
		// Use seq as conn key, map local port and remote port socket conn by seq
		newConn := new(Connection)
		newConn.Conn = conn
		newConn.Key = pkt.Seq
		CM.ConnMap[pkt.Seq] = newConn
		// Forward local to transfer forever
		errGup.Go(func() error {
			err := forwardToTransfer(newConn, clientObj.localPort)
			if err != nil {
				return err
			}
			return nil
		})
	}
	conn := CM.ConnMap[pkt.Seq]

	// Send payload to conn and send data back to conn
	errGup.Go(func() error {
		err := forwardToLocal(conn.Conn, pkt)
		if err != nil {
			return err
		}
		return nil
	})
	if err := errGup.Wait(); err != nil {
		return err
	}
	return nil
}

func handleFinalSignal(conn net.Conn, pkt *proto.NWPacket) error {
	return nil
}

func handleManager(mode string) map[uint8]Callback {
	handleMap := make(map[uint8]Callback)
	switch mode {
	case "server":
		handleMap[proto.FLG_REG] = handleRegistry
		handleMap[proto.FLG_HB] = handleHeartBeat
		handleMap[proto.FLG_FIN] = handleFinalSignal
		handleMap[proto.FLG_DAT] = handleDataServer
	case "client":
		handleMap[proto.FLG_REP] = handleReply
		handleMap[proto.FLG_DAT] = handleDataClient
		handleMap[proto.FLG_FIN] = handleFinalSignal
	}
	return handleMap
}
