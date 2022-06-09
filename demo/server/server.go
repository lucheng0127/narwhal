package main

import (
	"fmt"
	"io"
	"net"
	"sync"
)

var connMap = make(map[int]net.Conn)

func listenServer(port int) int {
	listen, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		fmt.Printf("Listen port %d error %s", port, err)
		panic(err)
	}
	fmt.Printf("Listen port %d\n", port)
	// Only accept single connection in demo
	conn, err := listen.Accept()
	if err != nil {
		fmt.Printf("Accept connection error %s", err)
	}
	rAddr := conn.RemoteAddr()
	rPort := rAddr.(*net.TCPAddr).Port
	fmt.Printf("Remote port %d\n", rPort)

	connMap[rPort] = conn
	return rPort
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
	// Listen 7023 wait connection
	transferPort := listenServer(7023)

	// Listen 80 wait connection, in demo we only handle one single connecion
	forwardPort := listenServer(4000)

	// Forward traffic between two connection
	var wg sync.WaitGroup
	go forwardTraffic(transferPort, forwardPort)
	go forwardTraffic(forwardPort, transferPort)
	wg.Add(2)
	wg.Wait()
}
