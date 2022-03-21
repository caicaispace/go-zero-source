package balancer

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"gozerosource/balancer/proto"
	"gozerosource/balancer/zrpc"

	"github.com/zeromicro/go-zero/core/discov"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
	"google.golang.org/grpc"
)

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

func (gs *BalancerServer) Hello(ctx context.Context, req *proto.Request) (*proto.Response, error) {
	fmt.Printf("im server 👉 %s => %s\n\n", time.Now().Format(timeFormat), req)

	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	return &proto.Response{
		Data: "hello from " + hostname,
	}, nil
}

func NewServer() {
	c := zrpc.RpcServerConf{
		ServiceConf: service.ServiceConf{
			Name: "rpc.balancer",
			Log: logx.LogConf{
				Mode: "console",
			},
		},
		ListenOn: "127.0.0.1:3456",
		Etcd: discov.EtcdConf{
			Hosts: []string{"127.0.0.1:2379"},
			Key:   "balancer.rpc",
		},
	}

	server := zrpc.MustNewServer(c, func(grpcServer *grpc.Server) {
		proto.RegisterBalancerServer(grpcServer, NewBalancerServer())
	})

	// 拦截器
	interceptor := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		st := time.Now()
		resp, err = handler(ctx, req)
		log.Printf("👋 method: %s time: %v\n\n", info.FullMethod, time.Since(st))
		return resp, err
	}

	server.AddUnaryInterceptors(interceptor)
	server.Start()
}

// func main() {
// 	NewServer()
// }
