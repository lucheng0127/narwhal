package protocol

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
