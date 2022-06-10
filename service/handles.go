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

type MOD int

const (
	MOD_T_S MOD = iota
	MOD_T_C
	MOD_S
	MOD_C
)

type Callback func(conn *Connection, pkt *proto.NWPacket) error
type mcCallback func(conn *Connection, mod int) error

// Monitor connection handlers
func handlePkt(transferConn *Connection, mod int) error {
	// TODO(lucheng): Fix handle pkt not forever

	// Check connection established before handle pkt
	// For server use the RemoteAddr port as key when listen server port,
	transferKey := transferConn.Conn.RemoteAddr().(*net.TCPAddr).Port
	if mod == 1 {
		transferKey = CM.ClientLocalPort
	}
	_, ok := CM.TransferConnMap[transferKey]
	if !ok {
		return new(internal.TransferConnNotExist)
	}
	handles := handleManager(mod)

	// Get pkt from transfer conn
	pkt, err := readFromTransferConn(transferConn)
	if pkt == nil {
		return nil
	}
	if err != nil {
		return err
	}
	if !pkt.Validate() {
		log.Debugf("Not a narwhal packet do nothing")
	}

	// Send to pkt handles
	errGup := new(errgroup.Group)
	errGup.Go(func() error {
		err := handles[pkt.Flag](transferConn, pkt)
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

func handleConn(conn *Connection, mod int) error {
	// Get transfer connection
	// For server use the LocalAddr port as key when listen target port,
	transferKey := conn.Conn.LocalAddr().(*net.TCPAddr).Port
	if mod == 1 {
		transferKey = CM.ClientLocalPort
	}
	_, ok := CM.TransferConnMap[transferKey]
	if !ok {
		log.Error("Transfer connection closed")
		return nil
	}
	transferConn := CM.TransferConnMap[transferKey]

	// Forward traffic to transferConn
	// datas from transfer conn will forward to conn in handle pkt
	errGup := new(errgroup.Group)
	errGup.Go(func() error {
		err := forwardToTransfer(conn, transferConn)
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

// Handlers for each narwhal packet

func handleRegistry(conn *Connection, pkt *proto.NWPacket) error {
	// Get target pot registry it
	targetPort := int(binary.BigEndian.Uint16(pkt.Payload))
	CM.Mux.Lock()
	// For server use the LocalAddr port as key when listen target port,
	// When receive registry request use target port as key
	CM.TransferConnMap[targetPort] = conn.Conn
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

	_, err = conn.Conn.Write(pktBytes)
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

func handleReply(conn *Connection, pkt *proto.NWPacket) error {
	if pkt.Result == proto.PORT_INUSE {
		panic("Target port used by others")
	}
	if pkt.Result != proto.RST_OK {
		panic("Registry failed")
	}
	log.Info("Registry succeed")
	return nil
}

func handleHeartBeat(conn *Connection, pkt *proto.NWPacket) error {
	return nil
}

func handleDataServer(conn *Connection, pkt *proto.NWPacket) error {
	log.Info(pkt.Seq)
	return nil
}

func handleDataClient(transferConn *Connection, pkt *proto.NWPacket) error {
	errGup := new(errgroup.Group)

	// Get seq num
	_, ok := CM.ConnMap[pkt.Seq]
	if !ok {
		// Dial to local port, add into ConnMap
		conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", CM.ClientLocalPort))
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
			err := forwardToTransfer(newConn, transferConn.Conn)
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

func handleFinalSignal(conn *Connection, pkt *proto.NWPacket) error {
	return nil
}

func handleManager(mode int) map[uint8]Callback {
	handleMap := make(map[uint8]Callback)
	switch mode {
	case 0:
		handleMap[proto.FLG_REG] = handleRegistry
		handleMap[proto.FLG_HB] = handleHeartBeat
		handleMap[proto.FLG_FIN] = handleFinalSignal
		handleMap[proto.FLG_DAT] = handleDataServer
	case 1:
		handleMap[proto.FLG_REP] = handleReply
		handleMap[proto.FLG_DAT] = handleDataClient
		handleMap[proto.FLG_FIN] = handleFinalSignal
	}
	return handleMap
}
