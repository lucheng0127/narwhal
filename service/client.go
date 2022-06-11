package service

import (
	"fmt"
	"narwhal/internal"
	"net"

	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

func handleForwardConn(conn *Connection) {
	for {
		// Fetch data to packet
		pktBytes, err := fetchDataToPktBytes(conn)
		if err != nil {
			panic(err)
		}

		// Send to transfer connection
		transferConn, ok := CM.TConnMap[CM.ServerAddr]
		if !ok {
			panic("Connection to narwhal server broken")
		}

		_, err = transferConn.Write(pktBytes)
		if err != nil {
			log.Errorf("Send packet bytes to transfer connection error\n%s", err.Error())
			continue
		}
	}
}

func handleClientConn(conn *Connection) error {
	handles := getHandles("Client")

	for {
		pkt, err := fetchPkt(conn)
		if err != nil {
			return err
		}
		if pkt == nil {
			// Connection closed out loop
			break
		}

		handles[pkt.Flag](pkt)
	}
	return nil
}

func RunClient(conf *internal.ClientConf) error {
	// Connection to narwhal
	serverAddr := fmt.Sprintf("%s:%d", conf.RemoteAddr, conf.ServerPort)
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		return internal.NewError("Connect to narwhal server error", err.Error())
	}
	CM.ClientLocalPort = conf.LocalPort
	CM.ServerAddr = conn.RemoteAddr().String()
	newConn := new(Connection)
	newConn.Conn = conn
	newConn.Key = newSeq()
	// No need send conn to ConnMap

	// Set TConnMap, serverAddr as key
	CM.Mux.Lock()
	CM.TConnMap[conn.RemoteAddr().String()] = conn
	CM.Mux.Unlock()
	errGup := new(errgroup.Group)
	log.Infof("New connection to %s local %s",
		newConn.Conn.RemoteAddr().String(), newConn.Conn.LocalAddr().String())

	// Groutine: Monitor conn and handle pkt
	errGup.Go(func() error {
		err := handleClientConn(newConn)
		if err != nil {
			return err
		}
		return nil
	})

	// TODO(lucheng): Registry client with target port

	if err := errGup.Wait(); err != nil {
		return err
	}
	return nil
}
