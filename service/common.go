package service

import (
	"math/rand"
	"narwhal/internal"
	"narwhal/proto"
	"net"

	log "github.com/sirupsen/logrus"
)

func newSeq() uint16 {
	seq := rand.Uint32() >> 16
	return uint16(seq)
}

func fetchDataToPktBytes(conn net.Conn, seq uint16) ([]byte, error) {
	// Read data from connection
	buf := make([]byte, proto.PayloadBufSize)
	n, err := conn.Read(buf)
	if err != nil {
		return nil, internal.NewError("Fetch data from proxy connection", err.Error())
	}

	// New narwhal packet
	pkt := new(proto.NWPacket)
	pkt.Flag = proto.FLG_DAT
	pkt.Seq = seq
	pkt.Result = proto.RST_OK
	pkt.SetPayload(buf[:n])
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

func fetchPkt(conn net.Conn) *proto.NWPacket {
	// Read data from connection
	buf := make([]byte, proto.BufSize)
	n, err := conn.Read(buf)
	if err != nil {
		panic(internal.NewError("Fetch packet from connection", err.Error()))
	}

	// Decode
	pkt, err := proto.Decode(buf[:n])
	if err != nil {
		panic(err)
	}
	log.Debugf("New packet received\nFlag: %d, Seq: %d", pkt.Flag, pkt.Seq)
	return pkt
}

func sendTofConn(pktChan chan *proto.NWPacket, conn net.Conn) {
	for {
		pkt := <-pktChan
		_, err := conn.Write(pkt.Payload)
		if err != nil {
			panic(err)
		}
	}
}

func sendTotConn(fConn, tConn net.Conn, seq uint16) {
	for {
		pktBytes, err := fetchDataToPktBytes(fConn, seq)
		if err != nil {
			log.Errorf("Fetch data to packet error, close it\n %s", err.Error())
			fConn.Close()
			return
		}

		_, err = tConn.Write(pktBytes)
		if err != nil {
			panic(err)
		}
	}
}

func switchTraffic(tConn, fConn net.Conn, pktChan chan *proto.NWPacket, seq uint16) {
	go sendTofConn(pktChan, fConn)
	go sendTotConn(fConn, tConn, seq)
}

func getOrCreatePktChan(seq uint16) chan *proto.NWPacket {
	pktChan, ok := client.pktChanMap[seq]
	if ok {
		return pktChan
	}

	client.mux.Lock()
	pktChan = make(chan *proto.NWPacket, 1024) // Maybe set it with config file, right now 1M is enough
	client.pktChanMap[seq] = pktChan
	client.mux.Unlock()

	return pktChan
}
