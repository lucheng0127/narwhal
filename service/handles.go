package service

import (
	"encoding/binary"
	"fmt"
	"narwhal/proto"
	"net"

	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type Callback func(conn net.Conn, pkt *proto.NWPacket) error

func listenLocal(port int) error {
	listen, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}
	// Add lister to lisMap
	serverCm.mux.Lock()
	targetPortLister := new(lister)
	targetPortLister.lister = listen
	targetPortLister.status = S_READY
	serverCm.lisMap[port] = targetPortLister
	serverCm.mux.Unlock()

	for {
		conn, err := listen.Accept()
		if err != nil {
			return err
		}
		// Add conn to connMap
		serverCm.mux.Lock()
		newConn := new(connection)
		newConn.conn = conn
		newConn.status = S_READY
		serverCm.connMap[conn.RemoteAddr().String()] = newConn
		serverCm.mux.Unlock()

		// Forward traffic
		for {
			err := forwardTraffic(&serverCm, conn)
			if err != nil {
				return err
			}
			return nil
		}
	}
}

func handleRegistry(conn net.Conn, pkt *proto.NWPacket) error {
	// Parase target port from payload, and mark transfer connection
	targetPort := int16(binary.BigEndian.Uint16(pkt.Payload))
	transferKey := fmt.Sprintf("targetPort-%d", int(targetPort))
	serverCm.mux.Lock()
	transferConn := new(connection)
	transferConn.conn = conn
	transferConn.status = S_DUBIOUS
	serverCm.connMap[transferKey] = transferConn
	serverCm.transferConnKey = transferKey
	serverCm.mux.Unlock()

	// Launch tcp server on localport
	errGroup := new(errgroup.Group)
	errGroup.Go(func() error {
		err := listenLocal(int(targetPort))
		if err != nil {
			return &hRegistryError{msg: err.Error()}
		}
		return nil
	})

	// Build reply packet
	repPkt := new(proto.NWPacket)
	repPkt.Flag = proto.FLG_REP
	repPkt.SetUnassignedAddrs()
	repPkt.Code = proto.C_OK
	repPkt.SetPayload([]byte("Registry reply packet"))
	err := repPkt.SetNoise()
	if err != nil {
		return &hRegistryError{msg: err.Error()}
	}
	repPktBytes, err := repPkt.Encode()
	if err != nil {
		return &hRegistryError{msg: err.Error()}
	}
	_, err = conn.Write(repPktBytes)
	if err != nil {
		return &hRegistryError{msg: err.Error()}
	}

	serverCm.mux.Lock()
	serverCm.connMap[transferKey].status = S_READY
	serverCm.mux.Unlock()
	log.Infof("Registry target port %d succeed", int(targetPort))

	// Listen port until error occor
	if err := errGroup.Wait(); err != nil {
		return err
	}
	return nil
}

func handleReply(conn net.Conn, pkt *proto.NWPacket) error {
	if pkt.Code != proto.C_OK {
		clientCm.mux.Lock()
		clientCm.connMap[clientCm.transferConnKey].status = S_DUBIOUS
		clientCm.mux.Unlock()
		log.Warn("Registry client failed")
	} else {
		clientCm.mux.Lock()
		clientCm.connMap[clientCm.transferConnKey].status = S_READY
		clientCm.mux.Unlock()
		log.Info("Registry client succeed")
	}
	return nil
}

func handleHeartBeat(conn net.Conn, pkt *proto.NWPacket) error {
	return nil
}

func handleDataServer(conn net.Conn, pkt *proto.NWPacket) error {
	return nil
}

func handleDataClient(conn net.Conn, pkt *proto.NWPacket) error {
	fmt.Printf("Pkt: %+v", pkt)
	log.Infof("Enter forward traffic wiating for implement")
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
	}
	return handleMap
}

func handlePkt(conn net.Conn, mod string) error {
	// Get server handles map
	handle := handleManager(mod)

	errGroup := new(errgroup.Group)

	// Fetch narwhal packet
	pkt, err := getPktFromConn(conn)
	if err != nil {
		return err
	}

	// Handle packet goroutine
	errGroup.Go(func() error {
		err := handle[pkt.Flag](conn, pkt)
		if err != nil {
			return err
		}
		return nil
	})
	return nil
}

func getPktFromConn(conn net.Conn) (*proto.NWPacket, error) {
	buf := make([]byte, proto.BufSize)
	_, err := conn.Read(buf)
	if err != nil {
		log.Warnf("Failed to read data from tcp connection %+v", conn)
		return nil, err
	}
	pkt, err := proto.Decode(buf)
	if err != nil {
		log.Errorf("Decode narwhal packet failed %s", err)
		return nil, err
	}
	return pkt, nil
}

func getPayloadFromConn(conn net.Conn) ([]byte, error) {
	buf := make([]byte, proto.PayloadBufSize)
	_, err := conn.Read(buf)
	if err != nil {
		return nil, &readError{msg: err.Error()}
	}
	return buf, nil
}

func netIPToUint32(ip net.IP) uint32 {
	if len(ip) == 16 {
		return binary.BigEndian.Uint32(ip[12:16])
	}
	return binary.BigEndian.Uint32(ip)
}

func CreatePacket(flag uint8, sAddr, cAddr string, pktBytes []byte) (*proto.NWPacket, error) {
	pkt := new(proto.NWPacket)

	// Reslove addr
	sAddrObj, err := net.ResolveTCPAddr("tcp", sAddr)
	if err != nil {
		log.Warnf("Failed to reslove tcp addr %s, set addr and port to zero", sAddr)
		pkt.SAddr = proto.UNASSIGNED_ADDR
		pkt.SPort = proto.UNASSIGNED_PORT
	} else {
		pkt.SAddr = netIPToUint32(sAddrObj.IP)
		pkt.SPort = uint16(sAddrObj.Port)
	}
	cAddrObj, err := net.ResolveTCPAddr("tcp", cAddr)
	if err != nil {
		log.Warnf("Failed to reslove tcp addr %s, set addr and port to zero", cAddr)
		pkt.CAddr = proto.UNASSIGNED_ADDR
		pkt.CPort = proto.UNASSIGNED_PORT
	} else {
		pkt.CAddr = netIPToUint32(cAddrObj.IP)
		pkt.CPort = uint16(cAddrObj.Port)
	}

	// Set flag payload and noise
	pkt.Flag = flag
	pkt.Code = proto.C_OK
	pkt.SetPayload(pktBytes)
	err = pkt.SetNoise()
	if err != nil {
		return nil, err
	}
	return pkt, nil
}
func forwardToNW(cm *connManager, conn net.Conn) error {
	// Forward traffic from socket, encode to
	// narwhal packet send to transfer socket

	// Read raw data from listen conn
	payload, err := getPayloadFromConn(conn)
	if err != nil {
		return err
	}

	// Create narwhal packet and encode
	pkt, err := CreatePacket(proto.FLG_DAT, conn.RemoteAddr().String(), proto.UNKNOWN_ADDR, payload)
	if err != nil {
		return err
	}
	pktBytes, err := pkt.Encode()
	if err != nil {
		return err
	}

	// Send narwhal packet to transferconn
	_, err = cm.connMap[cm.transferConnKey].conn.Write(pktBytes)
	if err != nil {
		return err
	}
	return nil
}

func Uint32ToIP(intIP uint32) net.IP {
	var IPBytes [4]byte
	IPBytes[0] = byte(intIP & 0xFF)
	IPBytes[1] = byte((intIP >> 8) & 0xFF)
	IPBytes[2] = byte((intIP >> 16) & 0xFF)
	IPBytes[3] = byte((intIP >> 24) & 0xFF)
	return net.IPv4(IPBytes[3], IPBytes[2], IPBytes[1], IPBytes[0])
}

func forwardToRaw(cm *connManager) error {
	// Read narwhal data from transferConn and decode
	pktBytes, err := getPayloadFromConn(cm.connMap[cm.transferConnKey].conn)
	if err != nil {
		return err
	}

	pkt, err := proto.Decode(pktBytes)
	if err != nil {
		return err
	}

	// Parse forwardConn addr from narwhal packet
	// if key not exist in connMap log it do not forward
	fAddr := Uint32ToIP(pkt.SAddr).String()
	fPort := int(pkt.SPort)
	fConnKey := fmt.Sprintf("%s:%d", fAddr, fPort)
	fConn, ok := cm.connMap[fConnKey]
	if !ok {
		return &connNotFound{msg: fConnKey}
	}

	_, err = fConn.conn.Write(pkt.Payload)
	if err != nil {
		return err
	}
	return nil
}

func forwardTraffic(cm *connManager, conn net.Conn) error {
	errGroup := new(errgroup.Group)
	errGroup.Go(func() error {
		err := forwardToNW(cm, conn)
		if err != nil {
			return err
		}
		return nil
	})
	errGroup.Go(func() error {
		err := forwardToRaw(cm)
		if err != nil {
			return err
		}
		return nil
	})
	// TODO(lucheng): Fix error not raise
	if err := errGroup.Wait(); err != nil {
		return err
	}
	return nil
}
