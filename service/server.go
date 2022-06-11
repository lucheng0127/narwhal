package service

import (
	"fmt"
	"narwhal/internal"
	"net"

	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
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
	//handles := serverHandle

	for {
		pkt, err := fetchPkt(conn)
		if err != nil {
			panic(err)
		}
		if pkt == nil {
			// Connection closed out loop
			break
		}

		go handleDataServer(pkt)
		// TODO(lucheng): Fix nil pointer for handles
		//go handles[pkt.Flag](pkt)
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
		// Add connection into ConnMap and TransferConnMap
		// For narwhal server use remote addr as key
		CM.Mux.Lock()
		CM.ConnMap[newConn.Key] = newConn
		//TODO(lucheng) Set TConnMap in registry
		CM.TConnMap["127.0.0.1:2222"] = conn
		CM.Mux.Unlock()
		// TODO(lucheng): Remove it
		NWServer.proxyMap[2222].tKey = conn.RemoteAddr().String()

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
		proxyAddr := conn.Conn.LocalAddr().String()
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
	errGup := new(errgroup.Group)

	// TODO(lucheng): Launch proxy server after registry
	proxyServer := new(ProxyServer)
	proxyServer.port = 2222
	NWServer.proxyMap[2222] = proxyServer

	errGup.Go(func() error {
		err := run(proxyServer)
		if err != nil {
			return err
		}
		return nil
	})

	err := run(&NWServer)
	if err != nil {
		return err
	}
	return nil
}
