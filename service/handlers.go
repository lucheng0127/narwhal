package service

import (
	"encoding/binary"
	"fmt"
	"narwhal/internal"
	"narwhal/proto"
	"net"

	log "github.com/sirupsen/logrus"
)

func handleReply(pkt *proto.NWPacket) {
	switch pkt.Result {
	case proto.RST_OK:
		log.Infof("Registry port %d succeed", client.rPort)
		return
	case proto.PORT_INUSE:
		eMsg := fmt.Sprintf("Registry port %d failed, port used", client.rPort)
		panic(eMsg)
	default:
		eMsg := fmt.Sprintf("Registry port %d failed", client.rPort)
		panic(eMsg)
	}
}

func handleRegistry(pkt *proto.NWPacket, conn net.Conn) {
	targetPort := int(binary.BigEndian.Uint16(pkt.Payload))

	// Launch proxy server
	pServer := new(proxyServer)
	pServer.port = targetPort
	server.mux.Lock()
	server.pServerMap[targetPort] = pServer
	server.tConnMap[targetPort] = conn
	server.mux.Unlock()

	err := run(pServer)

	// New reply pkt and send back
	repPkt := new(proto.NWPacket)
	repPkt.Flag = proto.FLG_REP
	repPkt.Seq = pkt.Seq
	if err == nil {
		repPkt.Result = proto.RST_OK
	} else if internal.IsPortInUsed(err) {
		repPkt.Result = proto.PORT_INUSE
	} else {
		repPkt.Result = proto.RST_ERR
	}
	repPkt.SetPayload(pkt.Payload)
	err = pkt.SetNoise()
	if err != nil {
		panic(internal.NewError("Reply regisgtry", err.Error()))
	}

	repPktBytes, err := repPkt.Encode()
	if err != nil {
		panic(internal.NewError("Reply regisgtry", err.Error()))
	}
	_, err = conn.Write(repPktBytes)
	if err != nil {
		panic(internal.NewError("Reply regisgtry", err.Error()))
	}
	log.Infof("Send reply packet for target port %d", targetPort)
}
