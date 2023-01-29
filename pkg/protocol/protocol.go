package protocol

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"net"

	logger "github.com/lucheng0127/narwhal/internal/pkg/log"
	"github.com/lucheng0127/narwhal/internal/pkg/utils"
)

// Request:
// +---+-------+
// |cmd|payload|
// +---+-------+
//
// auth: 0x8e
// bind: 0x8d
// new connection: 0x8b
// close: 0x87
// reply auth ctx: 0x89
// reply code: 0x88
//
// auth payload length 16 byte
// bind payload length 2 byte
// new connection payload length 16 byte, key of authedConn
// reply autx ctx payload length 16 byte
// reply code payload length 1 byte

const (
	// cmds
	CmdNone         = byte(0xcf)
	CmdAuth         = byte(0x8e)
	CmdBind         = byte(0x8d)
	CmdNewConn      = byte(0x8b)
	CmdClose        = byte(0x87)
	CmdReplyAuthCtx = byte(0x89)
	CmdReplyCode    = byte(0x88)

	// response code
	RepInvalidCmd  = byte(0xff)
	RepAuthFailed  = byte(0xfe)
	RepConnClose   = byte(0xfd)
	RepInvalidPort = byte(0xfb)
)

type PKG interface {
	Parse(ctx context.Context, conn net.Conn) error
	Encode() ([]byte, error)
	GetCmd() byte
	GetPayload() []byte
}

type RequestMethod struct {
	cmd     byte
	payload []byte
}

func NewRequestMethod(cmd byte, payload []byte) *RequestMethod {
	return &RequestMethod{cmd: cmd, payload: payload}
}

func (req *RequestMethod) GetCmd() byte {
	return req.cmd
}

func (req *RequestMethod) GetPayload() []byte {
	return req.payload
}

func (req *RequestMethod) Encode() ([]byte, error) {
	var tCtx utils.TraceCtx = utils.NewTraceID()
	ctx := tCtx.NewTraceContext()
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(req)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("encode request method cmd %s", err.Error()))
		return buf.Bytes(), err
	}
	return buf.Bytes(), nil
}

func (req *RequestMethod) Parse(ctx context.Context, conn net.Conn) error {
	// Parse method cmd
	methodBuf := make([]byte, 1)
	_, err := conn.Read(methodBuf)
	if err != nil {
		if err == io.EOF {
			return nil
		}
		return fmt.Errorf("parse request method cmd %s", err.Error())
	}

	// Parse method payload
	payloadLen := 0
	switch methodBuf[0] {
	case CmdAuth, CmdNewConn, CmdReplyAuthCtx:
		payloadLen = 16
	case CmdBind:
		payloadLen = 2
	case CmdClose:
		payloadLen = 0
	case CmdReplyCode:
		payloadLen = 1
	default:
		return errors.New("unsupport request method")
	}

	payloadBuf := make([]byte, payloadLen)
	n, err := conn.Read(payloadBuf)
	if err != nil && err != io.EOF {
		return err
	}
	if n != payloadLen {
		return errors.New("invalidate request method payload")
	}

	// Format req struct with methodBuf and payloadBuf
	req.cmd = methodBuf[0]
	req.payload = payloadBuf
	return nil
}
