package rpc

import (
	"fmt"
	pb "narwhal/proto"
	"net"
	"sync"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type RPCServer struct {
	pb.UnimplementedNarwhalServer
}

func LaunchRPCServer(rServer *RPCServer, port int, wg *sync.WaitGroup) {
	s := grpc.NewServer()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Errorf("Failed to listen port %d", port)
		wg.Done()
		panic(err)
	}
	pb.RegisterNarwhalServer(s, rServer)
	log.Infof("Running gRPC server at %v", lis.Addr())
	if err = s.Serve(lis); err != nil {
		log.Panic("Failed to server: %v", err)
		wg.Done()
		panic(err)
	}
}
