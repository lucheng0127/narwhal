package proto

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"net"
)

const (
	NWHeartbeat uint8 = 0xa1
	NWData      uint8 = 0xa3
	NWRegistry  uint8 = 0xa7
	NWReply     uint8 = 0xaf

	NWHeaderLen int = 5
	NoiseLen    int = 3
)

type Callback func(net.Conn, *NWPackage) error
type HandleMap map[uint8]Callback

type NWHeader struct {
	Flag   uint8
	Prefix uint16
	DLen   uint16
}

func (pkgHeader *NWHeader) String() string {
	flag := "Unknown"
	if pkgHeader.Flag^NWHeartbeat == 0 {
		flag = "Heartbeat"
	} else if pkgHeader.Flag^NWData == 0 {
		flag = "Data"
	} else if pkgHeader.Flag^NWRegistry == 0 {
		flag = "Registry"
	} else if pkgHeader.Flag^NWReply == 0 {
		flag = "Replay"
	}
	return fmt.Sprintf("Narwhal header: Flag %s, Prefix %d, DLen %d", flag, pkgHeader.Prefix, pkgHeader.DLen)
}

type NWPackage struct {
	NWHeader
	Payload []byte
	Noise   []byte
}

func (pkg *NWPackage) SetNoise() {
	// TODO:(lucheng)Minimum tcp data length 20 bytes
	minNoiseLen := 20 - NWHeaderLen - int(pkg.DLen)
	if minNoiseLen > NoiseLen {
		pkg.Noise = make([]byte, minNoiseLen)
	} else {
		pkg.Noise = make([]byte, NoiseLen)
	}
	rand.Read(pkg.Noise)
}

func (pkg *NWPackage) SetPayload(payload []byte) {
	pkg.Payload = payload
	pkg.DLen = uint16(len(pkg.Payload))
}

func (pkg *NWPackage) SetPrefix(prefix [2]byte) {
	pkg.Prefix = binary.BigEndian.Uint16(prefix[:])
}

func (pkg *NWPackage) Size() int {
	return NWHeaderLen + int(pkg.DLen) + len(pkg.Noise)
}

func (pkg *NWPackage) Marshal() ([]byte, error) {
	pkg.DLen = uint16(len(pkg.Payload))
	buf := bytes.NewBuffer(make([]byte, 0, pkg.Size()))
	// Write narwhal header into buf
	err := binary.Write(buf, binary.BigEndian, pkg.NWHeader)
	if err != nil {
		return nil, err
	}
	// Write payload into buf
	_, err = buf.Write(pkg.Payload)
	if err != nil {
		return nil, err
	}
	_, err = buf.Write(pkg.Noise)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (pkg *NWPackage) Unmarshal(b []byte) error {
	buf := bytes.NewBuffer(b)

	// Read data from buf and write it into pkg
	err := binary.Read(buf, binary.BigEndian, pkg.NWHeader)
	if err != nil {
		return err
	}
	pkg.Payload = make([]byte, pkg.DLen)
	// Read data from buf and wrie it into payload
	_, err = buf.Read(pkg.Payload)
	if err != nil {
		return err
	}
	return nil
}
