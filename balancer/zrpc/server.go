package zrpc

import (
	"log"
	"time"

	"gozerosource/balancer/zrpc/internal"
	"gozerosource/balancer/zrpc/internal/auth"
	"gozerosource/balancer/zrpc/internal/serverinterceptors"

	"github.com/zeromicro/go-zero/core/load"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stat"
	"google.golang.org/grpc"
)

// A RpcServer is a rpc server.
type RpcServer struct {
	server   internal.Server
	register internal.RegisterFn
}

// MustNewServer returns a RpcSever, exits on any error.
func MustNewServer(c RpcServerConf, register internal.RegisterFn) *RpcServer {
	server, err := NewServer(c, register)
	if err != nil {
		log.Fatal(err)
	}

	return server
}

// NewServer returns a RpcServer.
func NewServer(c RpcServerConf, register internal.RegisterFn) (*RpcServer, error) {
	var err error
	if err = c.Validate(); err != nil {
		return nil, err
	}

	var server internal.Server
	// 初始化服务指标监听器
	metrics := stat.NewMetrics(c.ListenOn)
	serverOptions := []internal.ServerOption{
		internal.WithMetrics(metrics),
	}

	if c.HasEtcd() {
		// 如果配置 etcd 服务则加载 rpc 发布服务 用于服务发现
		server, err = internal.NewRpcPubServer(c.Etcd, c.ListenOn, serverOptions...)
		if err != nil {
			return nil, err
		}
	} else {
		// 直接使用 rpc 服务
		server = internal.NewRpcServer(c.ListenOn, serverOptions...)
	}

	server.SetName(c.Name)
	if err = setupInterceptors(server, c, metrics); err != nil {
		return nil, err
	}

	rpcServer := &RpcServer{
		server:   server,
		register: register,
	}
	if err = c.SetUp(); err != nil {
		return nil, err
	}

	return rpcServer, nil
}

// AddOptions adds given options.
// 添加配置
func (rs *RpcServer) AddOptions(options ...grpc.ServerOption) {
	rs.server.AddOptions(options...)
}

// AddStreamInterceptors adds given stream interceptors.
// 添加 rpc 数据流拦截器
func (rs *RpcServer) AddStreamInterceptors(interceptors ...grpc.StreamServerInterceptor) {
	rs.server.AddStreamInterceptors(interceptors...)
}

// AddUnaryInterceptors adds given unary interceptors.
// 添加拦截器
func (rs *RpcServer) AddUnaryInterceptors(interceptors ...grpc.UnaryServerInterceptor) {
	rs.server.AddUnaryInterceptors(interceptors...)
}

// Start starts the RpcServer.
// Graceful shutdown is enabled by default.
// Use proc.SetTimeToForceQuit to customize the graceful shutdown period.
// 启动 Rpc 服务
// 已默认开启服务优雅关闭项
// 可使用 proc.SetTimeToForceQuit 配置项来自定义服务优雅关闭项
func (rs *RpcServer) Start() {
	if err := rs.server.Start(rs.register); err != nil {
		logx.Error(err)
		panic(err)
	}
}

// Stop stops the RpcServer.
func (rs *RpcServer) Stop() {
	logx.Close()
}

// SetServerSlowThreshold sets the slow threshold on server side.
// 设置服务器端的慢阈值
func SetServerSlowThreshold(threshold time.Duration) {
	serverinterceptors.SetSlowThreshold(threshold)
}

// 设置拦截器
func setupInterceptors(server internal.Server, c RpcServerConf, metrics *stat.Metrics) error {
	if c.CpuThreshold > 0 {
		// 添加服务降级拦截器
		shedder := load.NewAdaptiveShedder(load.WithCpuThreshold(c.CpuThreshold))
		server.AddUnaryInterceptors(serverinterceptors.UnarySheddingInterceptor(shedder, metrics))
	}

	if c.Timeout > 0 {
		// 添加服务超时拦截器
		server.AddUnaryInterceptors(serverinterceptors.UnaryTimeoutInterceptor(
			time.Duration(c.Timeout) * time.Millisecond))
	}

	if c.Auth {
		// 初始化权限验证服务
		authenticator, err := auth.NewAuthenticator(c.Redis.NewRedis(), c.Redis.Key, c.StrictControl)
		if err != nil {
			return err
		}

		// 添加 rpc 数据流拦截器
		server.AddStreamInterceptors(serverinterceptors.StreamAuthorizeInterceptor(authenticator))
		// 添加权限验证服务
		server.AddUnaryInterceptors(serverinterceptors.UnaryAuthorizeInterceptor(authenticator))
	}

	return nil
}
