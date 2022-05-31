package rpc

import (
	"context"
	pb "narwhal/proto"
	"time"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func SendHeartBeat(serverAddr string, port int) int {
	conn, err := grpc.Dial(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Errorf("Failed to connect to grpc server: %s, error:\n %s", serverAddr, err)
		return 1
	}
	defer conn.Close()

	c := pb.NewNarwhalClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	ret, err := c.HeartBeat(ctx, &pb.HBRequest{TargetPort: int32(port)})
	if err != nil {
		log.Errorf("Failed to send heart beat\n %s", err)
		return 1
	}
	log.Debugf("Heartbeat return code: %d", int(ret.Code))
	return int(ret.Code)
}

func (server *RPCServer) HeartBeat(ctx context.Context, in *pb.HBRequest) (*pb.HBReplay, error) {
	log.Debugf("Received heartbeat target port: %d", in.TargetPort)
	return &pb.HBReplay{Code: 0}, nil
}
