package proto

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"net"

	log "github.com/sirupsen/logrus"
)

func HandleHeartBeat(conn net.Conn, pkg *NWPackage) error {
	// TODO(lucheng): Implement it, mainten target port and connection map
	return nil
}

func GetHandles(mode string) HandleMap {
	if mode == "server" {
		return HandleMap{
			NWHeartbeat: HandleHeartBeat,
		}
	} else if mode == "client" {
		return HandleMap{}
	}
	return HandleMap{}
}

func SendHeartBeat(addr string, port int) error {
	serverAddr := fmt.Sprintf("%s:%d", addr, port)
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		return err
	}

	// Build narwhal package
	pkg := new(NWPackage)
	pkg.Flag = NWHeartbeat
	prefix := make([]byte, 2)
	var prefixArray [2]byte
	_, err = rand.Read(prefix)
	if err != nil {
		return err
	}
	_ = copy(prefixArray[:], prefix[:2])
	pkg.SetPrefix(prefixArray)
	payload := bytes.NewBuffer([]byte{})
	err = binary.Write(payload, binary.BigEndian, int16(port))
	if err != nil {
		return err
	}
	pkg.SetPayload(payload.Bytes())

	pkgBytes, err := pkg.Marshal()
	if err != nil {
		return err
	}
	n, err := conn.Write(pkgBytes)
	if err != nil {
		return err
	}
	log.Debugf("Send %d bytes to tcp conn:\nRemote info %+v Local info %+v", n, conn.RemoteAddr(), conn.LocalAddr())
	return nil
}
