package service

import (
	"narwhal/proto"
	"net"

	log "github.com/sirupsen/logrus"
)

type Callback func(conn net.Conn, pkt *proto.NWPacket) error

type connection struct {
	conn   net.Conn
	status string
	// TODO:(lucheng) Add lock
}

var connMap = make(map[int]connection)

func handleRegistry(conn net.Conn, pkt *proto.NWPacket) error {
	// Registry connection
	clientConn := new(connection)
	clientConn.conn = conn
	clientConn.status = "ESTABLISHED"
	connMap[int(pkt.TargetPort)] = *clientConn

	// Build reply packet
	repPkt := new(proto.NWPacket)
	repPkt.Flag = proto.FLG_REP
	repPkt.TargetPort = pkt.TargetPort
	repPkt.Option = proto.OPT_OK
	repPkt.SetPayload([]byte("Registry reply packet"))
	err := repPkt.SetNoise()
	if err != nil {
		log.Errorf("Set noise for reply packet error %s", err)
		return err
	}
	repPktBytes, err := repPkt.Encode()
	if err != nil {
		log.Errorf("Failed to encode reply packet %s", err)
		return err
	}
	_, err = conn.Write(repPktBytes)
	if err != nil {
		log.Errorf("Failed to reply client %s", err)
		return err
	}
	log.Infof("Registry target port %d succeed", int(pkt.TargetPort))
	return nil
}

func handleReply(conn net.Conn, pkt *proto.NWPacket) error {
	return nil
}

func handleHeartBeat(conn net.Conn, pkt *proto.NWPacket) error {
	return nil
}

func handleData(conn net.Conn, pkt *proto.NWPacket) error {
	return nil
}
func handleFinalSignal(conn net.Conn, pkt *proto.NWPacket) error {
	return nil
}

func handleManager(mode string) map[uint8]Callback {
	handleMap := make(map[uint8]Callback)
	handleMap[proto.FLG_DAT] = handleData
	switch mode {
	case "server":
		handleMap[proto.FLG_REG] = handleRegistry
		handleMap[proto.FLG_HB] = handleHeartBeat
		handleMap[proto.FLG_FIN] = handleFinalSignal
	case "client":
		handleMap[proto.FLG_REP] = handleReply
	}
	return handleMap
}

func getPktFromConn(conn net.Conn) (*proto.NWPacket, error) {
	buf := make([]byte, proto.BufSize)
	_, err := conn.Read(buf)
	if err != nil {
		log.Error("Failed to read data from tcp connection %+v", conn)
		return nil, err
	}
	pkt, err := proto.Decode(buf)
	if err != nil {
		log.Errorf("Decode narwhal packet failed %s", err)
		return nil, err
	}
	return pkt, nil
}
