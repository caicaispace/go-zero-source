package balancer

import (
	"context"
	"flag"
	"fmt"
	"time"

	"gozerosource/code/balancer/proto"
	"gozerosource/code/balancer/zrpc"

	"github.com/zeromicro/go-zero/core/discov"
)

const timeFormat = "15:04:05"

func NewClient() {
	flag.Parse()

	c := zrpc.RpcClientConf{
		Etcd: discov.EtcdConf{
			Hosts: []string{"127.0.0.1:2379"},
			Key:   "balancer.rpc",
		},
	}

	client := zrpc.MustNewClient(c)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			fmt.Println("")
			fmt.Println("---------------------------------------------")
			fmt.Println("")
			conn := client.Conn()
			balancer := proto.NewBalancerClient(conn)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			resp, err := balancer.Hello(ctx, &proto.Request{
				Msg: "im balancer test",
			})
			if err != nil {
				fmt.Printf("warning ⛔ %s X %s\n", time.Now().Format(timeFormat), err.Error())
			} else {
				fmt.Println("")
				fmt.Printf("im client 👉 %s => %s\n\n", time.Now().Format(timeFormat), resp.Data)
			}
			cancel()
		}
	}
}
