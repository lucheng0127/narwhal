package service

import (
	"fmt"
	"narwhal/internal"
	"narwhal/proto"
	"net"
	"runtime/debug"
	"sync"

	log "github.com/sirupsen/logrus"
)

type nwServer struct {
	mux        sync.RWMutex
	port       int
	tConnMap   map[int]net.Conn
	pServerMap map[int]*proxyServer
	//TODO(lucheng): make sure pkt orderly entry channel
	pktChanMap map[uint16]chan *proto.NWPacket
	pConnMap   map[uint16]net.Conn
}

type proxyServer struct {
	port   int
	lister net.Listener
}

var server = new(nwServer)

func init() {
	server.tConnMap = make(map[int]net.Conn)
	server.pServerMap = make(map[int]*proxyServer)
	server.pktChanMap = make(map[uint16]chan *proto.NWPacket)
	server.pConnMap = make(map[uint16]net.Conn)
}

func handleServerConn(conn net.Conn) {
	defer func() {
		if r := recover(); r != nil {
			log.Error("Handle connection local: %s, remote: %s error, close it\n%s",
				conn.LocalAddr().String(), conn.RemoteAddr().String(), string(debug.Stack()))
			conn.Close()
			return
		}
	}()

	// Read pkt from connection forever, send pkt to different handle according to flag
	for {
		pkt := fetchPkt(conn)

		switch pkt.Flag {
		case proto.FLG_DAT:
			_, ok := server.pConnMap[pkt.Seq]
			if !ok {
				log.Warn(fmt.Sprintf("No connection for seq %d, drop it", int(pkt.Seq)))
				continue
			}

			//  Append pkt to pkt chan
			pktChan := getOrCreatePktChan(pkt.Seq)
			pktChan <- pkt
		case proto.FLG_REG:
			go handleRegistry(pkt, conn)
		}
	}
}

func (s *nwServer) launch() error {
	lister, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return internal.NewError("TCP listen error", err.Error())
	}
	log.Infof("Narwhal server run on port %d", s.port)

	for {
		conn, err := lister.Accept()
		if err != nil {
			log.Error(internal.NewError("Establish new connection", err.Error()))
			continue
		}
		log.Infof(fmt.Sprintf("New connection local %s remote %s established",
			conn.LocalAddr().String(), conn.RemoteAddr().String()))

		// Monitor conn and handle pkt
		go handleServerConn(conn)
	}
}

func handleProxyConn(fConn net.Conn, seq uint16, pLister net.Listener) {
	server.mux.Lock()
	server.pConnMap[seq] = fConn
	server.mux.Unlock()

	targetPort := fConn.LocalAddr().(*net.TCPAddr).Port
	// Get transfer connection or close
	tConn, ok := server.tConnMap[targetPort]
	if !ok {
		log.Errorf("No transfer connection for port %d, shutdown proxy sever for it", targetPort)
		pLister.Close()
		tConn.Close()
		return
	}
	pktChan := getOrCreatePktChan(seq)

	// Switch traffic
	switchTraffic(tConn, fConn, pktChan, seq)
}

func (s *proxyServer) launch() error {
	// Close proxy server if error raised
	defer func() {
		if r := recover(); r != nil {
			log.Panicf("Proxy server error\n", debug.Stack())
			s.lister.Close()
		}
	}()

	lister, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return internal.NewError("Proxy listen error", err.Error())
	}
	s.lister = lister
	log.Infof("Proxy server run on port %d", s.port)

	go func() {
		for {
			conn, err := lister.Accept()
			if err != nil {
				lister.Close()
				log.Error(internal.NewError("Establish new connection", err.Error()).Error())
				continue
			}

			log.Infof("New connection to proxy server %d local: %s remote: %s",
				s.port, conn.LocalAddr().String(), conn.RemoteAddr().String())
			// Monitor conn forever read data encode it then send to transfer connection
			// Generate new seq for connection
			go handleProxyConn(conn, newSeq(), lister)
		}
	}()
	return nil
}

type tcpServer interface {
	launch() error
}

func run(server tcpServer) error {
	return server.launch()
}

func RunServer(conf *internal.ServerConf) error {
	log.Infof("Launch server with config: %+v", *conf)
	server.port = conf.ListenPort

	return run(server)
}
