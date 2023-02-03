package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"

	logger "github.com/lucheng0127/narwhal/internal/pkg/log"
	"github.com/lucheng0127/narwhal/internal/pkg/utils"
)

const (
	// Request code
	ReqNone   byte = byte(0xa0)
	ReqAuth   byte = byte(0xa1)
	ReqBind   byte = byte(0xa2)
	ReqPConn  byte = byte(0xa3) // Client establish a new connection with server send RepPConn to server with connection.AuthCtx
	ReqNotify byte = byte(0xa4) // A new connection establish to server binding port, server send RepNotify to client with connection.AuthCtx

	// Reply code
	RepNone   byte = byte(0x50)
	RepAuth   byte = byte(0x51)
	RepBind   byte = byte(0x52)
	RepPConn  byte = byte(0x53)
	RepNotify byte = byte(0x54)

	// Result code
	RetSucceed byte = byte(0xf0)
	RetFailed  byte = byte(0xf1)
)

// PKG is used to implement package for negotiation
//
// +-----+----+-------+
// |PCode|PLen|Payload|
// +-----+----+-------+
//
// PCode: request/reply method code
// PLen: length of payload
// Payload: payload of data
type PKG interface {
	Encode() ([]byte, error)
	SendToConn(conn net.Conn) error
	GetPCode() byte
	GetPayload() PL
}

type PL interface {
	String() string
	Int() int
}

type PHeader struct {
	PCode byte
	Plen  uint8
}

type PPayload struct {
	Data []byte
}

func (pp *PPayload) String() string {
	if len(pp.Data) == 0 {
		return ""
	}
	return string(pp.Data)
}

func (pp *PPayload) Int() int {
	if len(pp.Data) == 0 {
		return -1
	}
	return int(binary.BigEndian.Uint16(pp.Data))
}

type Package struct {
	Header  *PHeader
	Payload *PPayload
}

func NewPkt(code byte, payload []byte) PKG {
	pkt := new(Package)
	pkt.Header = new(PHeader)
	pkt.Payload = new(PPayload)
	pkt.Header.PCode = code
	pkt.Header.Plen = uint8(len(payload))
	pkt.Payload.Data = payload
	return pkt
}

func (p *Package) GetPCode() byte {
	return p.Header.PCode
}

func (p *Package) GetPayload() PL {
	return p.Payload
}

func ReadFromConn(conn net.Conn) (PKG, error) {
	pkt := new(Package)
	pkt.Header = new(PHeader)
	pkt.Payload = new(PPayload)

	err := binary.Read(conn, binary.BigEndian, pkt.Header)
	if err != nil {
		return nil, err
	}

	buf := make([]byte, int(pkt.Header.Plen))
	_, err = io.ReadFull(conn, buf)
	if err != nil {
		return nil, err
	}
	pkt.Payload.Data = buf[:]
	return pkt, nil
}

func (p *Package) SendToConn(conn net.Conn) error {
	pktBytes, err := p.Encode()
	if err != nil {
		return err
	}

	n, err := conn.Write(pktBytes)
	if err != nil {
		return err
	}

	ctx := utils.NewTraceContext()
	logger.Debug(ctx, fmt.Sprintf("sent [%d] bytes to connection %s", n, conn.RemoteAddr().String()))
	return nil
}

func (p *Package) Encode() ([]byte, error) {
	var headBuf bytes.Buffer

	err := binary.Write(&headBuf, binary.BigEndian, p.Header)
	if err != nil {
		return make([]byte, 0), nil
	}

	return append(headBuf.Bytes(), p.Payload.Data...), nil
}
