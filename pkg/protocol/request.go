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
	ctx := utils.NewTraceContext()
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
	//	// TODO(shawnlu): make req fix size
	//	err := binary.Read(conn, binary.BigEndian, req)
	//	if err != nil {
	//		return err
	//	}
	//	return nil

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
