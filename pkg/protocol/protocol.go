package protocol

import (
	"context"
	"net"
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
