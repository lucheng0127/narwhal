package connection

import "net"

type ConnPool struct {
	PoolMap map[string]net.Conn
	Pool    chan string
}

var GConnPool ConnPool

func init() {
	GConnPool := new(ConnPool)
	GConnPool.Pool = make(chan string, 32)
	GConnPool.PoolMap = make(map[string]net.Conn)
}
