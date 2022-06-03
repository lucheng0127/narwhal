package proto

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/rand"
)

const (
	// Narwhal header flags
	FLG_REG uint8 = 0x41 // Flag registry client
	FLG_REP uint8 = 0x51 // Flag registry reply
	FLG_HB  uint8 = 0x42 // Flag heartbeat
	FLG_DAT uint8 = 0x44 // Flag data
	FLG_FIN uint8 = 0x48 // Flag teardown client

	// Others
	NWHeaderLen       int = 6
	MinNoiseLen       int = 4
	MinimumPacketSize int = 20   // TCP minimum packet size
	BufSize           int = 1024 // Default layer3 MTU: 1500, less than 1500-20(IPHeaderlen)-20(TCPHeaderlen) is good
)

type NWHeader struct {
	Flag       uint8
	TargetPort uint16
	Length     uint16
	Option     uint8
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
	minNoiseLen := MinimumPacketSize - NWHeaderLen - int(pkt.Length)
	if minNoiseLen > 0 {
		pkt.Noise = make([]byte, minNoiseLen)
	} else {
		pkt.Noise = make([]byte, MinNoiseLen)
	}
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
