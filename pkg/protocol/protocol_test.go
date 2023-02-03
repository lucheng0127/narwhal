package protocol

import (
	"bytes"
	"encoding/binary"
	"net"
	"testing"
	"time"

	uuid "github.com/satori/go.uuid"
)

func TestSendAndReadPkt(t *testing.T) {
	uid := uuid.NewV4().String()
	type args struct {
		code    byte
		payload []byte
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "send and read",
			args: args{
				code:    RepAuth,
				payload: []byte(uid),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkt := NewPkt(tt.args.code, tt.args.payload)

			go func(pkt PKG) {
				ln, _ := net.Listen("tcp", "127.0.0.1:8881")
				for {
					conn, _ := ln.Accept()
					pkt.SendToConn(conn)
				}
			}(pkt)
			time.Sleep(100 * time.Microsecond)
			conn, _ := net.Dial("tcp", "127.0.0.1:8881")

			rPkt, err := ReadFromConn(conn)
			if err != nil {
				t.Error("read from connection error")
			}
			if rPkt.GetPCode() != RepAuth {
				t.Error("pcode not match")
			}
			if rPkt.GetPayload().String() != uid {
				t.Error("payload not match")
			}
		})
	}
}

func TestPPayload_Int(t *testing.T) {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, uint16(22))
	type fields struct {
		Data []byte
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{
			name:   "int -1",
			fields: fields{Data: make([]byte, 0)},
			want:   -1,
		},
		{
			name:   "int 22",
			fields: fields{Data: buf.Bytes()},
			want:   22,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pp := &PPayload{
				Data: tt.fields.Data,
			}
			if got := pp.Int(); got != tt.want {
				t.Errorf("PPayload.Int() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPPayload_String(t *testing.T) {
	type fields struct {
		Data []byte
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name:   "empty string",
			fields: fields{Data: make([]byte, 0)},
			want:   "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pp := &PPayload{
				Data: tt.fields.Data,
			}
			if got := pp.String(); got != tt.want {
				t.Errorf("PPayload.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
