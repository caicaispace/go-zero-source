package internal

import (
	"net"

	"gozerosource/balancer/zrpc/internal/serverinterceptors"

	"github.com/zeromicro/go-zero/core/proc"
	"github.com/zeromicro/go-zero/core/stat"
	"google.golang.org/grpc"
)

type (
	// ServerOption defines the method to customize a rpcServerOptions.
	ServerOption func(options *rpcServerOptions)

	rpcServerOptions struct {
		metrics *stat.Metrics
	}

	rpcServer struct {
		name string
		*baseRpcServer
	}
)

func init() {
	InitLogger()
}

// NewRpcServer returns a Server.
// 初始化 Rpc 服务
func NewRpcServer(address string, opts ...ServerOption) Server {
	var options rpcServerOptions
	for _, opt := range opts {
		opt(&options)
	}
	if options.metrics == nil {
		options.metrics = stat.NewMetrics(address)
	}

	return &rpcServer{
		baseRpcServer: newBaseRpcServer(address, &options),
	}
}

// 设置服务名称
func (s *rpcServer) SetName(name string) {
	s.name = name
	s.baseRpcServer.SetName(name)
}

// 启动服务
func (s *rpcServer) Start(register RegisterFn) error {
	lis, err := net.Listen("tcp", s.address)
	if err != nil {
		return err
	}

	// 基础拦截器
	unaryInterceptors := []grpc.UnaryServerInterceptor{
		serverinterceptors.UnaryTracingInterceptor,
		serverinterceptors.UnaryCrashInterceptor,
		serverinterceptors.UnaryStatInterceptor(s.metrics),
		serverinterceptors.UnaryPrometheusInterceptor,
		serverinterceptors.UnaryBreakerInterceptor,
	}
	unaryInterceptors = append(unaryInterceptors, s.unaryInterceptors...)
	// 基础流拦截器
	streamInterceptors := []grpc.StreamServerInterceptor{
		serverinterceptors.StreamTracingInterceptor,
		serverinterceptors.StreamCrashInterceptor,
		serverinterceptors.StreamBreakerInterceptor,
	}
	streamInterceptors = append(streamInterceptors, s.streamInterceptors...)
	options := append(s.options, WithUnaryServerInterceptors(unaryInterceptors...),
		WithStreamServerInterceptors(streamInterceptors...))
	server := grpc.NewServer(options...)
	register(server) // 加载业务服务
	// we need to make sure all others are wrapped up
	// so we do graceful stop at shutdown phase instead of wrap up phase
	// 确保所有其他服务都被包裹
	// 这样可以在关闭服务时对服务做优雅停止
	waitForCalled := proc.AddWrapUpListener(func() {
		server.GracefulStop()
	})
	defer waitForCalled()

	return server.Serve(lis)
}

// WithMetrics returns a func that sets metrics to a Server.
// 注册服务指标监听器
func WithMetrics(metrics *stat.Metrics) ServerOption {
	return func(options *rpcServerOptions) {
		options.metrics = metrics
	}
}
