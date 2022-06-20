package service

import (
	"io"
	"math/rand"
	"narwhal/internal"
	"narwhal/proto"

	log "github.com/sirupsen/logrus"
)

func newSeq() uint16 {
	seq := rand.Uint32() >> 16
	return uint16(seq)
}

func fetchDataToPktBytes(conn *Connection) ([]byte, error) {
	// Read data from connection
	buf := make([]byte, proto.PayloadBufSize)
	_, err := conn.Conn.Read(buf)
	if err != nil && err != io.EOF {
		return nil, internal.NewError("Fetch data from proxy connection", err.Error())
	} else if err == io.EOF {
		CM.Mux.Lock()
		conn.Status = C_CLOSED
		delete(CM.ConnMap, conn.Key)
		CM.Mux.Unlock()
		log.Warn("Proxy connection closed")
		return nil, nil
	}

	// New narwhal packet
	pkt := new(proto.NWPacket)
	pkt.Flag = proto.FLG_DAT
	pkt.Seq = conn.Key // Format seq with conn key
	pkt.Result = proto.RST_OK
	pkt.SetPayload(buf)
	err = pkt.SetNoise()
	if err != nil {
		return nil, err
	}

	// Return encode packet
	pktBytes, err := pkt.Encode()
	if err != nil {
		return nil, err
	}
	return pktBytes, nil
}

func fetchPkt(conn *Connection) (*proto.NWPacket, error) {
	// Read data from connection
	buf := make([]byte, proto.BufSize)
	_, err := conn.Conn.Read(buf)
	if err != nil && err != io.EOF {
		return nil, internal.NewError("Fetch packet from connection", err.Error())
	} else if err == io.EOF {
		log.Warn("Connection closed")
		return nil, nil
	}

	// Decode
	pkt, err := proto.Decode(buf)
	if err != nil {
		return nil, err
	}
	log.Debugf("New packet received\nFlag: %d, Seq: %d", pkt.Flag, pkt.Seq)
	return pkt, nil
}
