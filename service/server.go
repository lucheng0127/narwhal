package service

import (
	"fmt"
	"narwhal/internal"
	"narwhal/proto"
	"net"

	log "github.com/sirupsen/logrus"
)

type server interface {
	launch() error
}

func run(s server) error {
	return s.launch()
}

type NarwhalServer struct {
	port     int
	proxyMap map[int]*ProxyServer
}

var NWServer NarwhalServer

func init() {
	NWServer.proxyMap = make(map[int]*ProxyServer)
}

type ProxyServer struct {
	port int
	tKey string // Key to transfer connection for this porxy port, set it when target port registry
}

func handleServerConn(conn *Connection) {
	// Read pkt from connection forever, send pkt to different handle according to flag
	for {
		pkt, err := fetchPkt(conn)
		if err != nil {
			panic(err)
		}
		if pkt == nil {
			// Connection closed out loop
			break
		}

		switch pkt.Flag {
		case proto.FLG_DAT:
			go handleDataServer(pkt)
		case proto.FLG_REG:
			go handleRegistry(pkt, conn)
		}
	}
}

func (server *NarwhalServer) launch() error {
	lister, err := net.Listen("tcp", fmt.Sprintf(":%d", server.port))
	if err != nil {
		return internal.NewError("TCP listen error", err.Error())
	}
	log.Infof("Narwhal server run on port %d", server.port)

	for {
		conn, err := lister.Accept()
		if err != nil {
			return err
		}
		newConn := new(Connection)
		newConn.Key = newSeq()
		newConn.Conn = conn

		log.Infof("New connection to narwhal server remote: %s", newConn.Conn.RemoteAddr().String())
		// Monitor conn and handle pkt
		go handleServerConn(newConn)
	}
}

func handleProxyConn(conn *Connection) {
	for {
		// Fetch data to packet
		pktBytes, err := fetchDataToPktBytes(conn)
		if err != nil {
			panic(err)
		}

		// Send to transfer connection
		proxyAddr := fmt.Sprintf(":%d", conn.Conn.LocalAddr().(*net.TCPAddr).Port)
		transferConn, ok := CM.TConnMap[proxyAddr]
		if !ok {
			log.Errorf("Transfer connection for proxy %s closed", proxyAddr)
			break
		}
		n, err := transferConn.Write(pktBytes)
		if err != nil {
			log.Errorf("Send packet bytes to transfer connection error\n%s", err.Error())
			continue
		}
		log.Debugf("Send %d bytes to %s", n, transferConn.RemoteAddr().String())
	}
}

func (server *ProxyServer) launch() error {
	lister, err := net.Listen("tcp", fmt.Sprintf(":%d", server.port))
	if err != nil {
		return internal.NewError("Proxy listen error", err.Error())
	}
	log.Infof("Proxy server run on port %d", server.port)

	for {
		conn, err := lister.Accept()
		if err != nil {
			return err
		}
		newConn := new(Connection)
		newConn.Key = newSeq()
		newConn.Conn = conn
		// Add connection into ConnMap
		CM.Mux.Lock()
		CM.ConnMap[newConn.Key] = newConn
		CM.Mux.Unlock()

		log.Infof("New connection to proxy server %d remote: %s",
			server.port, newConn.Conn.RemoteAddr().String())
		// Monitor conn forever read data encode it then send to transfer connection
		go handleProxyConn(newConn)
	}
}

func RunServer(conf *internal.ServerConf) error {
	log.Infof("Launch server with config: %+v", *conf)
	NWServer.port = conf.ListenPort

	err := run(&NWServer)
	if err != nil {
		return err
	}
	return nil
}
