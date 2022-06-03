package service

import (
	"narwhal/proto"
	"net"
	"sync"

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
		log.Warnf("Failed to read data from tcp connection %+v", conn)
		return nil, err
	}
	return buf, nil
}

func forwardTrafficClient() error {
	// TODO(lucheng): Implement forward traffic for client
	// call forwardTraffic
	log.Infof("Enter forward traffic wiating for implement")
	return nil
}

func forwardToNW(targetPort int, transferConn, listenConn net.Conn) {
	// Read raw data from listen conn
	payload, err := getPayloadFromConn(listenConn)
	if err != nil {
		log.Debug(err)
	}

	// Create narwhal packet and encode
	pkt, err := proto.CreatePacket(targetPort, proto.FLG_DAT, payload)
	if err != nil {
		log.Warn(err)
	}
	pktBytes, err := pkt.Encode()
	if err != nil {
		log.Warn(err)
	}

	// Send narwhal packet to transferconn
	_, err = transferConn.Write(pktBytes)
	if err != nil {
		log.Warn(err)
	}
}

func forwardToRaw(transferConn, listenConn net.Conn) {
	// Read narwhal data from transferConn and decode
	pktBytes, err := getPayloadFromConn(transferConn)
	if err != nil {
		log.Warn(err)
	}
	pkt, err := proto.Decode(pktBytes)
	if err != nil {
		log.Warn(err)
	}

	// Send payload to listenConn
	_, err = listenConn.Write(pkt.Payload)
	if err != nil {
		log.Warn(err)
	}
}

func forwardTraffic(targetPort int, transferConn, listenConn net.Conn) {
	var wg sync.WaitGroup
	go forwardToNW(targetPort, transferConn, listenConn)
	go forwardToRaw(transferConn, listenConn)
	wg.Add(2)
	wg.Wait()
}
