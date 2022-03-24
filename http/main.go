package main

import (
	"fmt"
	"net/http"

	"gozerosource/http/rest"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
)

type ServiceContext struct {
	Config rest.RestConf
}

func NewServiceContext(c rest.RestConf) *ServiceContext {
	return &ServiceContext{
		Config: c,
	}
}

func IndexHandler(svcCtx *ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	}
}

func main() {
	c := rest.RestConf{
		Host:         "127.0.0.1",
		Port:         8081,
		MaxConns:     100,
		MaxBytes:     1048576,
		Timeout:      1000,
		CpuThreshold: 800,
		ServiceConf: service.ServiceConf{
			Log: logx.LogConf{
				Path: "./log",
			},
		},
	}
	server := rest.MustNewServer(c)
	defer server.Stop()
	ctx := NewServiceContext(c)
	server.AddRoute(rest.Route{
		Method:  http.MethodGet,
		Path:    "/ping",
		Handler: IndexHandler(ctx),
	})

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
