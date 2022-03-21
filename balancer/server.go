package balancer

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"gozerosource/balancer/pb"
	"gozerosource/balancer/zrpc"

	"github.com/zeromicro/go-zero/core/conf"
	"google.golang.org/grpc"
)

var serverConfigFile = flag.String("f", "config.json", "the config file")

type BalancerServer struct {
	lock     sync.Mutex
	alive    bool
	downTime time.Time
}

func NewBalancerServer() *BalancerServer {
	return &BalancerServer{
		alive: true,
	}
}

func (gs *BalancerServer) Hello(ctx context.Context, req *pb.Request) (*pb.Response, error) {
	fmt.Println("=>", req)

	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	return &pb.Response{
		Data: "hello from " + hostname,
	}, nil
}

func NewServer() {
	flag.Parse()

	var c zrpc.RpcServerConf
	conf.MustLoad(*serverConfigFile, &c)

	server := zrpc.MustNewServer(c, func(grpcServer *grpc.Server) {
		pb.RegisterBalancerServer(grpcServer, NewBalancerServer())
	})
	interceptor := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		st := time.Now()
		resp, err = handler(ctx, req)
		log.Printf("method: %s time: %v\n", info.FullMethod, time.Since(st))
		return resp, err
	}

	server.AddUnaryInterceptors(interceptor)
	server.Start()
}

// func main() {
// 	NewServer()
// }
