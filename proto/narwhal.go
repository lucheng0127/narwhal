package proto

import (
	"bytes"
	"encoding/binary"
	"math/rand"
	"narwhal/internal"
)

const (
	// Narwhal header flags
	FLG_REG     uint8 = 0x41 // Flag registry client
	FLG_REP     uint8 = 0x51 // Flag registry reply
	FLG_HB      uint8 = 0x42 // Flag heartbeat
	FLG_DAT     uint8 = 0x44 // Flag data
	FLG_FIN     uint8 = 0x48 // Flag teardown client
	FLG_FIN_REP uint8 = 0x58 // Flag teardown reply
	// Result code
	RST_OK     uint8 = 0xa0 // Indicate result correct
	RST_ERR    uint8 = 0xa1 // Indicate result error occur
	PORT_INUSE uint8 = 0xa2 // Target port used by other

	// Others
	MinTCPPktLen   int = 20
	NWHeaderLen    int = 6
	MinNoiseLen    int = 4    // Minium noise length, make sure packet large than 20 bytes
	BufSize        int = 1034 // 1034 + 20(TCPHeader) + 20(IPHeader) < 1500
	PayloadBufSize int = 1024

	// Net addr
	UNKNOWN_ADDR string = "Unknown addr" // Unknown address, set addr and port to zero
)

type NWHeader struct {
	Flag   uint8  // Indicate packet type
	Seq    uint16 // Seq num, also key of connMap
	Length uint16 // Payload length
	Result uint8  // Result of request type packets
}

type NWPacket struct {
	NWHeader
	Payload []byte
	Noise   []byte
}

func (pkt *NWPacket) Validate() bool {
	switch pkt.Flag {
	case FLG_REG, FLG_REP, FLG_HB, FLG_DAT, FLG_FIN, FLG_FIN_REP:
		return true
	default:
		return false
	}
}

func (pkt *NWPacket) SetNoise() error {
	minNoiseLen := MinTCPPktLen - NWHeaderLen - int(pkt.Length)
	if minNoiseLen > MinNoiseLen {
		pkt.Noise = make([]byte, minNoiseLen)
	} else {
		pkt.Noise = make([]byte, MinNoiseLen)
	}
	_, err := rand.Read(pkt.Noise)
	if err != nil {
		return internal.NewError("Protocol set noise", err.Error())
	}
	return nil
}

func (pkt *NWPacket) SetPayload(b []byte) {
	pkt.Payload = b
	pkt.Length = uint16(len(pkt.Payload))
}

func (pkt *NWPacket) SetTargetPort(port int16) error {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, int16(port))
	if err != nil {
		return internal.NewError("Protocol set target port", err.Error())
	}
	pkt.SetPayload(buf.Bytes())
	return nil
}

func (pkt *NWPacket) Size() int {
	return NWHeaderLen + len(pkt.Payload) + len(pkt.Noise)
}

func (pkt *NWPacket) Encode() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, pkt.NWHeader)
	if err != nil {
		return nil, internal.NewError("Protocol encode header", err.Error())
	}
	_, err = buf.Write(pkt.Payload)
	if err != nil {
		return nil, internal.NewError("Protocol set payload", err.Error())
	}
	_, err = buf.Write(pkt.Noise)
	if err != nil {
		return nil, internal.NewError("Protocol write noise ", err.Error())
	}
	return buf.Bytes(), nil
}

func Decode(b []byte) (*NWPacket, error) {
	pkt := new(NWPacket)
	buf := bytes.NewReader(b)
	err := binary.Read(buf, binary.BigEndian, &pkt.NWHeader)
	if err != nil {
		return nil, internal.NewError("Protocol decode header", err.Error())
	}

	pkt.Payload = make([]byte, pkt.Length)
	_, err = buf.Read(pkt.Payload)
	if err != nil {
		return nil, internal.NewError("Protocol read payload", err.Error())
	}
	return pkt, nil
}
