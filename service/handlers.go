package service

import (
	"fmt"
	"narwhal/internal"
	"narwhal/proto"
	"net"

	log "github.com/sirupsen/logrus"
)

type pktHandle func(pkt *proto.NWPacket)

var serverHandle map[uint8]pktHandle
var clientHandle map[uint8]pktHandle

func init() {
	// Init server handles
	serverHandle := make(map[uint8]pktHandle)
	serverHandle[proto.FLG_DAT] = handleDataServer

	// Init client handles
	clientHandle := make(map[uint8]pktHandle)
	clientHandle[proto.FLG_DAT] = handleDataClient
}

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
	_, err := forwardConn.Conn.Write(pkt.Payload)
	if err != nil {
		log.Errorf("Send data to connection %s failed\n%s",
			forwardConn.Conn.RemoteAddr().String(), err.Error())
	}
}

// Narwhal server handlers
func handleDataServer(pkt *proto.NWPacket) {
	// Get conn via pkt.Seq
	targetConn, ok := CM.ConnMap[pkt.Seq]
	if !ok {
		log.Error("Connection for seq %d closed", int(pkt.Seq))
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

// Narwhal client handlers

// Proxy server handlers
