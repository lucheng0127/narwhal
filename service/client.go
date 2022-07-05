package service

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"narwhal/internal"
	"narwhal/proto"
	"net"
	"runtime/debug"
	"sync"

	log "github.com/sirupsen/logrus"
)

type nwClient struct {
	mux        sync.Mutex
	lPort      int
	rPort      int
	serverAddr string
	pktChanMap map[uint16]chan *proto.NWPacket
	fConnMap   map[uint16]net.Conn
	tConn      net.Conn
}

var client nwClient

func init() {
	client.pktChanMap = make(map[uint16]chan *proto.NWPacket)
	client.fConnMap = make(map[uint16]net.Conn)
}

func registryTargetPort(conn net.Conn, targetPort int) error {
	// Send registry packet
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, uint16(targetPort))
	if err != nil {
		return internal.NewError("Registry client", err.Error())
	}

	// New narwhal packet
	pkt := new(proto.NWPacket)
	pkt.Flag = proto.FLG_REG
	pkt.Seq = newSeq()
	pkt.Result = proto.RST_OK
	pkt.SetPayload(buf.Bytes())
	err = pkt.SetNoise()
	if err != nil {
		return internal.NewError("Registry client", err.Error())
	}

	pktBytes, err := pkt.Encode()
	if err != nil {
		return internal.NewError("Registry client", err.Error())
	}

	// Send reg pkt, if registry failed panic in handle reply
	_, err = conn.Write(pktBytes)
	if err != nil {
		return internal.NewError("Registry client", err.Error())
	}
	log.Debugf("Send registry pkt for target port %d", targetPort)
	return nil
}

func getOrNewConn(seq uint16) net.Conn {
	conn, ok := client.fConnMap[seq]
	if ok {
		return conn
	}

	conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", client.lPort))
	if err != nil {
		panic(fmt.Sprintf("Conect to local port %d\n%s", client.lPort, err.Error()))
	}

	return conn
}

func handleSeqClient(seq uint16) {
	// Get or create connection to local port
	fConn := getOrNewConn(seq)
	pktChan := getOrCreatePktChan(seq)

	// Switch traffic
	switchTraffic(client.tConn, fConn, pktChan, seq)
}

func handleClientConn(conn net.Conn, wg sync.WaitGroup) {
	defer func() {
		if r := recover(); r != nil {
			log.Panicf("Client error\n", string(debug.Stack()))
			conn.Close()
			wg.Done()
		}
	}()

	for {
		pkt := fetchPkt(conn)

		switch pkt.Flag {
		case proto.FLG_DAT:
			go handleSeqClient(pkt.Seq)

			// Get or create pktChan
			pktChan := getOrCreatePktChan(pkt.Seq)

			// Append pkt to pktChan
			pktChan <- pkt
		case proto.FLG_REP:
			handleReply(pkt)
		}
	}
}

func RunClient(conf *internal.ClientConf) error {
	// Connection to narwhal
	var wg sync.WaitGroup
	client.lPort = conf.LocalPort
	client.rPort = conf.RemotePort
	client.serverAddr = fmt.Sprintf("%s:%d", conf.RemoteAddr, conf.ServerPort)

	conn, err := net.Dial("tcp", client.serverAddr)
	if err != nil {
		return internal.NewError("Connect to narwhal server error", err.Error())
	}
	client.tConn = conn
	log.Infof("Connect to server local: %s, remote: %s",
		client.tConn.LocalAddr().String(), client.tConn.RemoteAddr().String())

	// Groutine: Monitor conn and handle pkt
	go handleClientConn(client.tConn, wg)
	wg.Add(1)

	// Registry client with target port
	err = registryTargetPort(client.tConn, client.rPort)
	if err != nil {
		return err
	}

	wg.Wait()
	return nil
}
