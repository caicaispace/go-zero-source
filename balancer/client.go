package balancer

import (
	"context"
	"flag"
	"fmt"
	"time"

	"gozerosource/balancer/pb"
	"gozerosource/balancer/zrpc"

	"github.com/zeromicro/go-zero/core/conf"
)

const timeFormat = "15:04:05"

var configFile = flag.String("f", "config.yaml", "the config file")

func NewClient() {
	flag.Parse()

	var c zrpc.RpcClientConf
	conf.MustLoad(*configFile, &c)

	client := zrpc.MustNewClient(c)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			conn := client.Conn()
			balancer := pb.NewBalancerClient(conn)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			resp, err := balancer.Hello(ctx, &pb.Request{
				Name: "kevin",
			})
			if err != nil {
				fmt.Printf("%s X %s\n", time.Now().Format(timeFormat), err.Error())
			} else {
				fmt.Printf("%s => %s\n", time.Now().Format(timeFormat), resp.Data)
			}
			cancel()
		}
	}
}
