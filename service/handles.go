package service

import (
	"encoding/binary"
	"fmt"
	"io"
	"narwhal/proto"
	"net"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type Callback func(conn net.Conn, pkt *proto.NWPacket) error

// Functions called by handlers

func listenTargetPort(port int) error {
	// Server listen target port, hand new connection,
	// forward traffic come from target port to transfer conneciton
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

		// Forward traffic for new connection
		errGroup := new(errgroup.Group)
		for {
			errGroup.Go(func() error {
				err := forwardToNW(&serverCm, conn, conn.RemoteAddr().String(), proto.UNKNOWN_ADDR)
				if err != nil {
					return err
				}
				return nil
			})
			if err := errGroup.Wait(); err != nil {
				panic(err)
			}
		}
	}
}

func monitorLocalPort(conn net.Conn, sAddr, cAddr string) error {
	// Monitor local port traffic from connection
	// forward traffic from conn to transfer connection
	for {
		err := forwardToNW(&clientCm, conn, sAddr, cAddr)
		if err != nil {
			return err
		}
	}
}

// Handlers for each narwhal packet

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
		err := listenTargetPort(int(targetPort))
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
	fmt.Printf("%+v", pkt)
	return nil
}

func handleDataClient(transferConn net.Conn, pkt *proto.NWPacket) error {
	// Get connMap key, from SAddr and SPort
	sAddr := Uint32ToIP(pkt.SAddr)
	connKey := fmt.Sprintf("%s:%d", sAddr.String(), int(pkt.SPort))

	// If conn not exist, try to connect to local port,
	// add conn to connMap, set connection.peerAddr
	errGroup := new(errgroup.Group)
	_, ok := clientCm.connMap[connKey]
	if !ok {
		localPortString := strings.Split(clientCm.transferConnKey, "-")
		localPort, err := strconv.Atoi(localPortString[1])
		if err != nil {
			return err
		}
		newConn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", localPort))
		if err != nil {
			panic(err)
		}
		log.Debugf("New connection local address %s remote address %s",
			newConn.LocalAddr().String(), newConn.RemoteAddr().String())
		conn := new(connection)
		conn.conn = newConn
		conn.status = S_READY
		conn.peerAddr = connKey
		clientCm.connMap[connKey] = conn

		// Start a new goroutine read data from conn send back to transferConn
		errGroup.Go(func() error {
			err := monitorLocalPort(newConn, connKey, newConn.LocalAddr().String())
			if err != nil {
				return err
			}
			return nil
		})
	}

	// Forward traffic between local socket and transfer socket
	log.Debugf("Send %d bytes data to %s", len(pkt.Payload), clientCm.connMap[connKey].conn.RemoteAddr().String())
	_, err := clientCm.connMap[connKey].conn.Write(pkt.Payload)
	if err != nil {
		return err
	}
	if err := errGroup.Wait(); err != nil {
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

// Utils functions for connection and packets

func Uint32ToIP(intIP uint32) net.IP {
	var IPBytes [4]byte
	IPBytes[0] = byte(intIP & 0xFF)
	IPBytes[1] = byte((intIP >> 8) & 0xFF)
	IPBytes[2] = byte((intIP >> 16) & 0xFF)
	IPBytes[3] = byte((intIP >> 24) & 0xFF)
	return net.IPv4(IPBytes[3], IPBytes[2], IPBytes[1], IPBytes[0])
}

func getPktFromConn(conn net.Conn) (*proto.NWPacket, error) {
	buf := make([]byte, proto.BufSize)
	_, err := conn.Read(buf)
	if err != nil && err != io.EOF {
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
	if err != nil && err != io.EOF {
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

// Forward functions

func forwardToNW(cm *connManager, conn net.Conn, sAddr, cAddr string) error {
	// Forward traffic from socket, encode to
	// narwhal packet send to transfer socket

	// Read raw data from listen conn
	payload, err := getPayloadFromConn(conn)
	if err != nil {
		return err
	}
	log.Debugf("Read %d bytes data from %s", len(payload), conn.RemoteAddr().String())

	// Create narwhal packet and encode
	pkt, err := CreatePacket(proto.FLG_DAT, sAddr, cAddr, payload)
	if err != nil {
		return err
	}
	pktBytes, err := pkt.Encode()
	if err != nil {
		return err
	}

	// Send narwhal packet to transferconn
	// TODO(lucheng): Fix send data error, connection broken
	log.Debugf("Try to send data to connection remote %s, local %s",
		cm.connMap[cm.transferConnKey].conn.RemoteAddr().String(),
		cm.connMap[cm.transferConnKey].conn.LocalAddr().String())
	_, err = cm.connMap[cm.transferConnKey].conn.Write(pktBytes)
	if err != nil {
		log.Error(err)
		return err
	}
	log.Debugf("Send %d bytes data to %s", len(pktBytes), cm.connMap[cm.transferConnKey].conn.RemoteAddr().String())
	return nil
}
