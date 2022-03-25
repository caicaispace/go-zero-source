package rest

import (
	"fmt"
	"net/http"

	"gozerosource/code/rest/rest"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
)

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
	server.AddRoutes([]rest.Route{
		{
			Method: http.MethodGet,
			Path:   "/ping",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("pong"))
			},
		},
		{
			Method: http.MethodGet,
			Path:   "/check",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("ok"))
			},
		},
	})

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
