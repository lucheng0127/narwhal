package protocol

import (
	"context"
	"errors"
	"fmt"
	"net"

	logger "github.com/lucheng0127/narwhal/internal/pkg/log"
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
//
// auth payload length 16 byte
// bind payload length 2 byte
// new connection payload length 16 byte, key of authedConn

const (
	// cmds
	CmdAuth    = byte(0x8e)
	CmdBind    = byte(0x8d)
	CmdNewConn = byte(0x8b)
	CmdClose   = byte(0x87)

	// response code
	RepInvalidCmd  = byte(0xff)
	RepAuthFailed  = byte(0xfe)
	RepConnClose   = byte(0xfd)
	RepInvalidPort = byte(0xfb)
)

type PKG interface {
	Parse(ctx context.Context, conn net.Conn) error
	GetCmd() byte
	GetPayload() []byte
}

type RequestMethod struct {
	Cmd     byte
	payload []byte
}

func NewRequestMethod() *RequestMethod {
	return new(RequestMethod)
}

func (req *RequestMethod) GetCmd() byte {
	return req.Cmd
}

func (req *RequestMethod) GetPayload() []byte {
	return req.payload
}

func (req *RequestMethod) Parse(ctx context.Context, conn net.Conn) error {
	// Parse method cmd
	methodBuf := make([]byte, 1)
	_, err := conn.Read(methodBuf)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("parse request method cmd %s", err.Error()))
		return err
	}

	// Parse method payload
	payloadLen := 0
	switch methodBuf[0] {
	case CmdAuth:
		payloadLen = 16
	case CmdBind:
		payloadLen = 2
	case CmdClose:
		payloadLen = 0
	default:
		logger.Error(ctx, "unsupport request method")
		return errors.New("unsupport request method")
	}

	payloadBuf := make([]byte, payloadLen)
	n, err := conn.Read(payloadBuf)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("parse request method payload %s", err.Error()))
		return err
	}
	if n != payloadLen {
		logger.Error(ctx, "invalidate request method payload")
		return errors.New("invalidate request method payload")
	}

	// Format req struct with methodBuf and payloadBuf
	req.Cmd = methodBuf[0]
	req.payload = payloadBuf
	return nil
}
