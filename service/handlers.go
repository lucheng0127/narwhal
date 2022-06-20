package service

import (
	"encoding/binary"
	"fmt"
	"narwhal/internal"
	"narwhal/proto"
	"net"

	log "github.com/sirupsen/logrus"
)

// Narwhal client handlers
func handleDataClient(pkt *proto.NWPacket) {
	// Get connection from pkt.Seq
	_, ok := CM.ConnMap[pkt.Seq]
	if !ok {
		// Create if not exist
		// Connect to local port
		conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", CM.ClientLocalPort))
		if err != nil {
			panic(internal.NewError("Connect to forward port", err.Error()))
		}
		newConn := new(Connection)
		newConn.Conn = conn
		newConn.Key = pkt.Seq
		// Add connection into ConnMap
		CM.Mux.Lock()
		CM.ConnMap[newConn.Key] = newConn
		CM.Mux.Unlock()
	}
	// Reget connection
	forwardConn := CM.ConnMap[pkt.Seq]

	// Groutine: Monitor conn forever
	go handleForwardConn(forwardConn)

	// Send pkt.Payload to conn
	// TODO(lucheng): Send the whole traffic data to connection without split it
	// Like ssh preauth packet, packet size large than 1024, if split it into
	// several parts then send to ssh connection, connection will reset by peer,
	// becaues of message authentication code incorrect
	_, err := forwardConn.Conn.Write(pkt.Payload)
	if err != nil {
		log.Errorf("Send data to connection %s failed\n%s",
			forwardConn.Conn.RemoteAddr().String(), err.Error())
	}
}

func handleReply(pkt *proto.NWPacket) {
	switch pkt.Result {
	case proto.RST_OK:
		log.Infof("Registry port %d succeed", CM.ClientLocalPort)
		return
	default:
		eMsg := fmt.Sprintf("Registry port %d failed", CM.ClientLocalPort)
		panic(eMsg)
	}
}

// Narwhal server handlers
func handleDataServer(pkt *proto.NWPacket) {
	// Get conn via pkt.Seq
	targetConn, ok := CM.ConnMap[pkt.Seq]
	if !ok {
		log.Errorf("Connection for seq %d closed", int(pkt.Seq))
		return
	}

	// Send pkt.Payload to conn
	n, err := targetConn.Conn.Write(pkt.Payload)
	if err != nil {
		log.Error("Send data to connection %s failed\n%s",
			targetConn.Conn.RemoteAddr().String(), err.Error())
	}
	log.Debugf("Send %d bytes data to connection %s",
		n, targetConn.Conn.RemoteAddr().String())
}

func handleRegistry(pkt *proto.NWPacket, conn *Connection) {
	targetPort := int(binary.BigEndian.Uint16(pkt.Payload))

	// Launch proxy server
	proxyServer := new(ProxyServer)
	proxyServer.port = targetPort
	NWServer.proxyMap[targetPort] = proxyServer

	go run(proxyServer)

	// New reply pkt and send back
	repPkt := new(proto.NWPacket)
	repPkt.Flag = proto.FLG_REP
	repPkt.Seq = conn.Key
	repPkt.Result = proto.RST_OK
	repPkt.SetPayload(pkt.Payload)
	err := pkt.SetNoise()
	if err != nil {
		panic(internal.NewError("Reply regisgtry", err.Error()))
	}
	// Store transfer connection for target port
	CM.Mux.Lock()
	CM.TConnMap[fmt.Sprintf(":%d", targetPort)] = conn.Conn
	CM.Mux.Unlock()

	repPktBytes, err := repPkt.Encode()
	if err != nil {
		panic(internal.NewError("Reply regisgtry", err.Error()))
	}
	_, err = conn.Conn.Write(repPktBytes)
	if err != nil {
		panic(internal.NewError("Reply regisgtry", err.Error()))
	}
	log.Infof("Send reply packet for target port %d", targetPort)
}

// Narwhal client handlers

// Proxy server handlers
