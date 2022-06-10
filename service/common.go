package service

import (
	"fmt"
	"io"
	"math/rand"
	"narwhal/internal"
	"narwhal/proto"
	"net"

	log "github.com/sirupsen/logrus"
)

// Utils funcs

func newNWPkt(flag uint8, seq uint16, payload []byte) (*proto.NWPacket, error) {
	pkt := new(proto.NWPacket)
	pkt.Flag = flag
	pkt.Seq = seq
	pkt.Result = proto.RST_OK
	pkt.SetPayload(payload)
	err := pkt.SetNoise()
	if err != nil {
		return nil, err
	}
	return pkt, nil
}

func readFromLocalConn(conn net.Conn) ([]byte, error) {
	buf := make([]byte, proto.PayloadBufSize)
	n, err := conn.Read(buf)
	if err != nil && err != io.EOF {
		return nil, internal.NewError("Read local connection error", err.Error())
	}
	log.Debugf("Read %d bytes from local connection", n)
	return buf, nil
}

func forwardToTransfer(conn *Connection, transferConn net.Conn) error {
	localConn := conn.Conn

	for {
		// Read data from local conn and encode
		buf, err := readFromLocalConn(localConn)
		if err != nil {
			return internal.NewError("Read local connection error", err.Error())
		}

		// Encode to nw packet
		// Use conn key as seq num, so when packet back can find conn to porxy port
		pkt, err := newNWPkt(proto.FLG_DAT, conn.Key, buf)
		if err != nil {
			return err
		}
		pktBytes, err := pkt.Encode()
		if err != nil {
			return err
		}

		// Write to transfer conn
		n, err := transferConn.Write(pktBytes)
		if err != nil {
			return internal.NewError("Send data to transfer connection error", err.Error())
		}
		log.Debugf("Send %d bytes data to transfer connection, seq %d", n, int(pkt.Seq))
	}
}

func readFromTransferConn(transferConn *Connection) (*proto.NWPacket, error) {
	buf := make([]byte, proto.BufSize)
	n, err := transferConn.Conn.Read(buf)
	if err == io.EOF {
		// TODO(lucheng): client should exit program
		log.Warn("Connection closed by client")
		// Rmove conn from TransferConnMap and ConnMap
		// For client transferConn key is CM.ClientLocalPort
		// For server transferConn key is remote addr
		_, ok := CM.TransferConnMap[CM.ClientLocalPort]
		if ok {
			delete(CM.TransferConnMap, CM.ClientLocalPort)
		}
		_, ok = CM.TransferConnMap[transferConn.Conn.RemoteAddr().(*net.TCPAddr).Port]
		if ok {
			delete(CM.TransferConnMap, transferConn.Conn.RemoteAddr().(*net.TCPAddr).Port)
		}
		_, ok = CM.ConnMap[transferConn.Key]
		if ok {
			delete(CM.ConnMap, transferConn.Key)
		}
		return nil, nil
	} else if err != nil {
		return nil, internal.NewError("Read transfer connection error", err.Error())
	}

	pkt, err := proto.Decode(buf)
	if err != nil {
		return nil, err
	}

	log.Debugf("Read %d bytes from transfer connection, seq %d", n, pkt.Seq)
	return pkt, nil
}

func forwardToLocal(conn net.Conn, pkt *proto.NWPacket) error {
	// Write to local conn
	n, err := conn.Write(pkt.Payload)
	if err != nil {
		return internal.NewError("Send data to local connection error", err.Error())
	}
	log.Debugf("Send %d bytes data to local connection", n)
	return nil
}

func newTCPServer(port int) (net.Listener, error) {
	lister, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, internal.NewError("TCP listen error", err.Error())
	}
	return lister, nil
}

func serviceTCPServer(lister net.Listener) (*Connection, error) {
	conn, err := lister.Accept()
	if err != nil {
		return nil, internal.NewError("TCP accept connection error", err.Error())
	}
	// Add conn into ConnMap
	newConn := new(Connection)
	newConn.Conn = conn
	newConn.Key = newSeq()
	CM.Mux.Lock()
	CM.ConnMap[newConn.Key] = newConn
	CM.Mux.Unlock()
	log.Debugf("New connection local %s, remotes %s\nseq %d",
		newConn.Conn.LocalAddr().String(), newConn.Conn.RemoteAddr().String(),
		newConn.Key)
	return newConn, nil
}

func newSeq() uint16 {
	seq := rand.Uint32() >> 16
	return uint16(seq)
}

func listenAndService(port int) error {
	// Launch proxy port TCP server, call by handle registry
	lister, err := newTCPServer(port)
	if err != nil {
		return err
	}

	for {
		conn, err := serviceTCPServer(lister)
		if err != nil {
			// TODO(lucheng): handle port in used
			return err
		}

		// Run forward grountine
		go monitorConn(conn, int(MOD_S))
	}
}

func doMonitorConn(conn *Connection, mod int, callback mcCallback) {
	for {
		err := callback(conn, mod)
		if err != nil {
			if internal.IsConnClosed(err) {
				// Conn closed return
				break
			}
			panic(err)
		}
	}
}

func monitorConn(conn *Connection, mod int) {
	// Mointor connections,
	// For transfer connections use different handle to handle traffic
	// For connection connect with internet and target port, use forwardTraffic
	// to forward traffic between transfer connection and connection
	// For connection connect to client local port, use forwardTraffic
	// to forward traffic between transfer connection and connection
	switch mod {
	case 0, 1:
		doMonitorConn(conn, mod, handlePkt)
	case 2, 3:
		doMonitorConn(conn, mod, handleConn)
	}
}
