package main

import (
	"fmt"
	"io"
	"net"
	"sync"
)

var connMap = make(map[int]net.Conn)

func dialServer(port int) int {
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		panic(err)
	}

	lAddr := conn.LocalAddr()
	lPort := lAddr.(*net.TCPAddr).Port
	fmt.Printf("Connect to server with local port %d\n", lPort)
	connMap[lPort] = conn
	return lPort
}

func forwardTraffic(srcConnPort, dstConnPort int) {
	for {
		buf := make([]byte, 1024)
		n, err := connMap[srcConnPort].Read(buf)
		if err != nil && err != io.EOF {
			panic(err)
		}
		connMap[dstConnPort].Write(buf[:n])
	}
}

func main() {
	// Dial 7023
	transferPort := dialServer(7023)

	// Dial 8000
	forwardPort := dialServer(22)

	// Forward traffic
	var wg sync.WaitGroup
	go forwardTraffic(transferPort, forwardPort)
	go forwardTraffic(forwardPort, transferPort)
	wg.Add(2)
	wg.Wait()
}
