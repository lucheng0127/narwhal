package protocol

import "net"

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
	Decode() error
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
	return ""
}

func (pp *PPayload) Int() int {
	return -1
}

type Package struct {
	Header  *PHeader
	Payload *PPayload
}

func NewPkt(code byte, payload []byte) PKG {
	pkt := new(Package)
	pkt.Header.PCode = code
	pkt.Header.Plen = uint8(len(payload))
	pkt.Payload.Data = payload
	return pkt
}

func ReadFromConn(conn net.Conn) (PKG, error) {
	// TODO
	pkt := new(Package)
	pkt.Header = new(PHeader)
	pkt.Payload = new(PPayload)
	return pkt, nil
}

func (p *Package) SendToConn(conn net.Conn) error {
	return nil
}

func (p *Package) Encode() ([]byte, error) {
	return make([]byte, 0), nil
}

func (p *Package) Decode() error {
	return nil
}

func (p *Package) GetPCode() byte {
	return p.Header.PCode
}

func (p *Package) GetPayload() PL {
	return p.Payload
}
