package internal

import (
	"github.com/zeromicro/go-zero/core/stat"
	"google.golang.org/grpc"
)

type (
	// RegisterFn defines the method to register a server.
	// 定义注册服务器的方法，用于加载业务服务
	RegisterFn func(*grpc.Server)

	// Server interface represents a rpc server.
	// 服务接口
	Server interface {
		// 添加配置
		AddOptions(options ...grpc.ServerOption)
		// 添加 rpc 数据流拦截器
		AddStreamInterceptors(interceptors ...grpc.StreamServerInterceptor)
		// 添加拦截器
		AddUnaryInterceptors(interceptors ...grpc.UnaryServerInterceptor)
		// 设置服务名称
		SetName(string)
		// 启动服务
		Start(register RegisterFn) error
	}

	// rpc 服务基类
	baseRpcServer struct {
		address            string
		metrics            *stat.Metrics
		options            []grpc.ServerOption
		streamInterceptors []grpc.StreamServerInterceptor
		unaryInterceptors  []grpc.UnaryServerInterceptor
	}
)

// 初始化 rpc 服务基类
func newBaseRpcServer(address string, rpcServerOpts *rpcServerOptions) *baseRpcServer {
	return &baseRpcServer{
		address: address,
		metrics: rpcServerOpts.metrics,
	}
}

func (s *baseRpcServer) AddOptions(options ...grpc.ServerOption) {
	s.options = append(s.options, options...)
}

func (s *baseRpcServer) AddStreamInterceptors(interceptors ...grpc.StreamServerInterceptor) {
	s.streamInterceptors = append(s.streamInterceptors, interceptors...)
}

func (s *baseRpcServer) AddUnaryInterceptors(interceptors ...grpc.UnaryServerInterceptor) {
	s.unaryInterceptors = append(s.unaryInterceptors, interceptors...)
}

func (s *baseRpcServer) SetName(name string) {
	s.metrics.SetName(name)
}
