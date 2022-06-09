package service

import (
	"narwhal/proto"
	"net"
)

type Callback func(conn net.Conn, pkt *proto.NWPacket) error

// Handlers for each narwhal packet

func handleRegistry(conn net.Conn, pkt *proto.NWPacket) error {
	return nil
}

func handleReply(conn net.Conn, pkt *proto.NWPacket) error {
	return nil
}

func handleHeartBeat(conn net.Conn, pkt *proto.NWPacket) error {
	return nil
}

func handleDataServer(conn net.Conn, pkt *proto.NWPacket) error {
	return nil
}

func handleDataClient(transferConn net.Conn, pkt *proto.NWPacket) error {
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
