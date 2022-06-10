package service

import (
	"narwhal/internal"
	"net"

	log "github.com/sirupsen/logrus"
)

type server struct {
	serverPort int
}

var serverObj server

func serviceNWServer(lister net.Listener) (*Connection, error) {
	// Accept new connection to server port
	// store connection to TransferConnMap. target port as key
	// this connection use to transfer traffic from the internet
	// to target port, each connection connect with internet and
	// target port use a unique seq num
	conn, err := lister.Accept()
	if err != nil {
		return nil, internal.NewError("Narwhal server accept connection error", err.Error())
	}
	log.Debugf("New connection established, remote address %s", conn.RemoteAddr().String())

	// Add connection to transferConnMap
	CM.Mux.Lock()
	newConn := new(Connection)
	newConn.Conn = conn
	newConn.Key = newSeq()
	CM.ConnMap[newConn.Key] = newConn
	CM.TransferConnMap[conn.RemoteAddr().(*net.TCPAddr).Port] = conn
	CM.Mux.Unlock()
	return newConn, nil
}

func launchNWServer(port int) error {
	// Listen servr port, waiting for registry request
	// or transfer data, mark with seq num, each seq num
	// indicate an socket connection stored in ConnMap
	// After registry a target port, use target port
	// as key of TransferConnMap, connection as value
	lister, err := newTCPServer(port)
	if err != nil {
		return err
	}

	// Accept connection forever
	for {
		conn, err := serviceNWServer(lister)
		if err != nil {
			return err
		}

		// Monitor packet from transfer connection
		// narwhal registry or transfer data
		go monitorConn(conn, int(MOD_T_S))
	}
}

func RunServer(conf *internal.ServerConf) error {
	log.Infof("Launch server with config: %+v", *conf)
	serverObj.serverPort = conf.ListenPort

	// Launch tcp server and listen forever
	// in it handle new connection
	err := launchNWServer(serverObj.serverPort)
	if err != nil {
		return err
	}
	return nil
}
