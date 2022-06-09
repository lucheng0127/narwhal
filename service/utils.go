package service

import (
	"fmt"
	"io"
	"math/rand"
	"narwhal/internal"
	"narwhal/proto"
	"net"

	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
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

func forwardToTransfer(conn *Connection, transferKey int) error {
	transferConn := CM.TransferConnMap[transferKey]
	localConn := conn.Conn

	for {
		// Read data from local conn and encode
		buf, err := readFromLocalConn(localConn)
		if err != nil {
			return internal.NewError("Read local connection error", err.Error())
		}

		// Encode to nw packet
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
		log.Debugf("Send %d bytes data to transfer connection", n)
	}
}

func readFromTransferConn(transferKey int) (*proto.NWPacket, error) {
	buf := make([]byte, proto.BufSize)
	n, err := CM.TransferConnMap[transferKey].Read(buf)
	if err == io.EOF {
		log.Warn("Connection closed by client")
		// Rmove conn from TransferConnMap
		delete(CM.TransferConnMap, transferKey)
		return nil, nil
	} else if err != nil {
		return nil, internal.NewError("Read transfer connection error", err.Error())
	}
	log.Debugf("Read %d bytes from transfer connection", n)

	pkt, err := proto.Decode(buf)
	if err != nil {
		return nil, err
	}

	return pkt, nil
}

func newConnToLocal(seq uint16) error {
	// Dial to local TCP server
	conn, err := net.Dial("tcp", fmt.Sprintf(":%d", clientObj.localPort))
	if err != nil {
		return err
	}
	newConn := new(Connection)
	newConn.Conn = conn
	newConn.Key = seq
	// Add to ConnMap
	CM.Mux.Lock()
	CM.ConnMap[seq] = newConn
	CM.Mux.Unlock()
	return nil
}

func forwardToLocal(transferKey int) error {
	// Parse narwhal packet from transfer conn, get seq
	pkt, err := readFromTransferConn(transferKey)
	if pkt == nil {
		return nil
	}
	if !pkt.Validate() {
		log.Debugf("Not a narwhal packet, do nothing")
		return nil
	}
	if err != nil {
		return err
	}

	// Get local conn from ConnMap
	// Create one if not exist
	_, ok := CM.ConnMap[pkt.Seq]
	if !ok {
		err := newConnToLocal(pkt.Seq)
		if err != nil {
			return internal.NewError("Connection to local error", err.Error())
		}
	}
	localConn := CM.ConnMap[pkt.Seq]

	// Write to local conn
	n, err := localConn.Conn.Write(pkt.Payload)
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
			return err
		}

		// Run forward grountine
		errGup := new(errgroup.Group)
		errGup.Go(func() error {
			transferKey := conn.Conn.LocalAddr().(*net.TCPAddr).Port
			err := forwardToTransfer(conn, transferKey)
			if err != nil {
				return err
			}
			return nil
		})

		if err = errGup.Wait(); err != nil {
			return err
		}
	}
}

func serviceNWServer(lister net.Listener) (net.Conn, error) {
	// Accept connection
	conn, err := lister.Accept()
	if err != nil {
		return nil, internal.NewError("Narwhal server accept connection error", err.Error())
	}
	log.Debugf("New connection established, remote address %s", conn.RemoteAddr().String())

	// Add connection to transferConnMap
	CM.Mux.Lock()
	CM.TransferConnMap[conn.RemoteAddr().(*net.TCPAddr).Port] = conn
	CM.Mux.Unlock()
	return conn, nil
}

func _handPkt(conn net.Conn, mod string, transferKey int) error {
	// Check transferKey exist before handPkt
	_, ok := CM.TransferConnMap[transferKey]
	if !ok {
		return new(internal.TransferConnNotExist)
	}
	handles := handleManager(mod)

	// Get pkt from transfer conn
	pkt, err := readFromTransferConn(transferKey)
	if pkt == nil {
		return nil
	}
	if !pkt.Validate() {
		log.Debugf("Not a narwhal packet, do nothing")
		return nil
	}
	if err != nil {
		return err
	}

	errGup := new(errgroup.Group)
	errGup.Go(func() error {
		err := handles[pkt.Flag](conn, pkt)
		if err != nil {
			return err
		}
		return nil
	})
	if err := errGup.Wait(); err != nil {
		return err
	}
	return nil
}

func handPkt(conn net.Conn, mod string) {
	for {
		transferKey := conn.RemoteAddr().(*net.TCPAddr).Port
		err := _handPkt(conn, "server", transferKey)
		if err != nil {
			if internal.IsConnClosed(err) {
				// Conn closed return
				break
			}
			panic(err)
		}
	}
}

func launchNWServer(port int) error {
	// Listen servr port and registry transfer conn
	lister, err := newTCPServer(port)
	if err != nil {
		return err
	}

	// Accept connection forever
	for {
		conn, err := serviceNWServer(lister)
		if err != nil {
			return err
		}

		// Handle pkt forever with groutine, until conn closed
		go handPkt(conn, "server")
	}
}
