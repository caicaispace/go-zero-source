package main

import (
	"fmt"
	"net/http"

	"gozerosource/rest/rest"

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

func PingHandler(svcCtx *ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	}
}

func CheckHandler(svcCtx *ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}
}

func ServerStart() {
	c := rest.RestConf{
		Host:         "127.0.0.1",
		Port:         8081,
		MaxConns:     100,
		MaxBytes:     1048576,
		Timeout:      1000,
		CpuThreshold: 800,
		ServiceConf: service.ServiceConf{
			Log: logx.LogConf{
				Mode: "console",
				Path: "./logs",
			},
		},
	}
	server := rest.MustNewServer(c, rest.WithCors("localhost:8080"))
	defer server.Stop()
	ctx := NewServiceContext(c)
	server.AddRoutes([]rest.Route{
		{
			Method:  http.MethodGet,
			Path:    "/ping",
			Handler: PingHandler(ctx),
		},
		{
			Method:  http.MethodGet,
			Path:    "/check",
			Handler: CheckHandler(ctx),
		},
	})

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}

func main() {
	ServerStart()
}
