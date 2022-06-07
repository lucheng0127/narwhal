package proto

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/rand"
)

const (
	// Narwhal header flags
	FLG_REG     uint8 = 0x41 // Flag registry client
	FLG_REP     uint8 = 0x51 // Flag registry reply
	FLG_HB      uint8 = 0x42 // Flag heartbeat
	FLG_DAT     uint8 = 0x44 // Flag data
	FLG_FIN     uint8 = 0x48 // Flag teardown client
	FLG_FIN_REP uint8 = 0x58 // Flag teardown reply
	// Narwhal addr and port
	UNASSIGNED_ADDR uint32 = 0x0 // Unassigned address
	UNASSIGNED_PORT uint16 = 0x0 // Unassigned port
	// Option result
	C_OK  uint8 = 0xa0 // Option indecate result correct
	C_ERR uint8 = 0xa1 // Option indecate result error occur

	// Others
	NWHeaderLen    int = 16
	NoiseLen       int = 4    // Minium tcp packet length 20, NWHeaderLen + NoiseLen = 20 it ok send empty packet
	BufSize        int = 1024 // Default layer3 MTU: 1500, less than 1500-20(IPHeaderlen)-20(TCPHeaderlen) is good
	PayloadBufSize int = 1004 // Bufsize of narwhal size is 1024, payload max size = 1024 - 16 - 4

	// Net addr
	UNKNOWN_ADDR string = "Unknown addr" // Unknown address, set addr and port to zero
)

type NWHeader struct {
	Flag   uint8  // Indicate packet type
	SAddr  uint32 // Socket address of server that communicate to target port
	SPort  uint16 // Socket port of server that communicate to target port
	CAddr  uint32 // Socket address of client that communicate to forward port
	CPort  uint16 // Socket port of client that communicate to forward port
	Length uint16 // Payload length
	Code   uint8  // Used by reply type packet, OK or ERR to indicate request type packet handle succeed or failed
}

type NWPacket struct {
	NWHeader
	Payload []byte
	Noise   []byte
}

type ProtoError struct {
	msg string
}

type ProtoEncodeError struct {
	ProtoError
}

type ProtoDecodeError struct {
	ProtoError
}

func (err *ProtoError) Error() string {
	return fmt.Sprintf("Narwhal protocol error %s", err.msg)
}

func (err *ProtoEncodeError) Error() string {
	return fmt.Sprintf("Narwhal protocol encode error %s", err.msg)
}

func (err *ProtoDecodeError) Error() string {
	return fmt.Sprintf("Narwhal protocol decode error %s", err.msg)
}

func (pkt *NWPacket) SetNoise() error {
	pkt.Noise = make([]byte, NoiseLen)
	_, err := rand.Read(pkt.Noise)
	if err != nil {
		return &ProtoError{msg: err.Error()}
	}
	return nil
}

func (pkt *NWPacket) SetPayload(b []byte) {
	pkt.Payload = b
	pkt.Length = uint16(len(pkt.Payload))
}

func (pkt *NWPacket) SetUnassignedAddrs() {
	pkt.SAddr = UNASSIGNED_ADDR
	pkt.SPort = UNASSIGNED_PORT
	pkt.CAddr = UNASSIGNED_ADDR
	pkt.CPort = UNASSIGNED_PORT
}

func (pkt *NWPacket) SetTargetPort(port int16) error {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, int16(port))
	if err != nil {
		return &ProtoError{msg: err.Error()}
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
		return nil, &ProtoEncodeError{ProtoError{msg: err.Error()}}
	}
	_, err = buf.Write(pkt.Payload)
	if err != nil {
		return nil, &ProtoEncodeError{ProtoError{msg: err.Error()}}
	}
	_, err = buf.Write(pkt.Noise)
	if err != nil {
		return nil, &ProtoEncodeError{ProtoError{msg: err.Error()}}
	}
	return buf.Bytes(), nil
}

func Decode(b []byte) (*NWPacket, error) {
	pkt := new(NWPacket)
	buf := bytes.NewReader(b)
	err := binary.Read(buf, binary.BigEndian, &pkt.NWHeader)
	if err != nil {
		return nil, &ProtoDecodeError{ProtoError{msg: err.Error()}}
	}

	pkt.Payload = make([]byte, pkt.Length)
	_, err = buf.Read(pkt.Payload)
	if err != nil {
		return nil, &ProtoDecodeError{ProtoError{msg: err.Error()}}
	}
	return pkt, nil
}
