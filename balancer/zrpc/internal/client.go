package internal

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gozerosource/balancer/zrpc/internal/clientinterceptors"
	"gozerosource/balancer/zrpc/p2c"
	"gozerosource/balancer/zrpc/resolver"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	dialTimeout = time.Second * 3
	separator   = '/'
)

func init() {
	resolver.Register()
}

type (
	// 客户端接口
	// Client interface wraps the Conn method.
	Client interface {
		Conn() *grpc.ClientConn
	}

	// 客户端配置项
	// A ClientOptions is a client options.
	ClientOptions struct {
		NonBlock    bool
		Timeout     time.Duration
		Secure      bool
		DialOptions []grpc.DialOption
	}

	// ClientOption defines the method to customize a ClientOptions.
	ClientOption func(options *ClientOptions)

	client struct {
		conn *grpc.ClientConn
	}
)

// 初始化客户端
// NewClient returns a Client.
func NewClient(target string, opts ...ClientOption) (Client, error) {
	var cli client
	opts = append([]ClientOption{WithDialOption(grpc.WithBalancerName(p2c.Name))}, opts...)
	if err := cli.dial(target, opts...); err != nil {
		return nil, err
	}

	return &cli, nil
}

// 获取客户端连接
func (c *client) Conn() *grpc.ClientConn {
	return c.conn
}

// 构建拨号配置
func (c *client) buildDialOptions(opts ...ClientOption) []grpc.DialOption {
	var cliOpts ClientOptions
	for _, opt := range opts {
		opt(&cliOpts)
	}

	var options []grpc.DialOption
	if !cliOpts.Secure {
		options = append([]grpc.DialOption(nil), grpc.WithInsecure())
	}

	if !cliOpts.NonBlock {
		options = append(options, grpc.WithBlock())
	}

	options = append(options,
		WithUnaryClientInterceptors(
			clientinterceptors.UnaryTracingInterceptor,
			clientinterceptors.DurationInterceptor,
			clientinterceptors.PrometheusInterceptor,
			clientinterceptors.BreakerInterceptor,
			clientinterceptors.TimeoutInterceptor(cliOpts.Timeout),
		),
		WithStreamClientInterceptors(
			clientinterceptors.StreamTracingInterceptor,
		),
	)

	return append(options, cliOpts.DialOptions...)
}

// 拨号
func (c *client) dial(server string, opts ...ClientOption) error {
	options := c.buildDialOptions(opts...)
	timeCtx, cancel := context.WithTimeout(context.Background(), dialTimeout)
	defer cancel()
	conn, err := grpc.DialContext(timeCtx, server, options...)
	if err != nil {
		service := server
		if errors.Is(err, context.DeadlineExceeded) {
			pos := strings.LastIndexByte(server, separator)
			// len(server) - 1 is the index of last char
			if 0 < pos && pos < len(server)-1 {
				service = server[pos+1:]
			}
		}
		return fmt.Errorf("rpc dial: %s, error: %s, make sure rpc service %q is already started",
			server, err.Error(), service)
	}

	c.conn = conn
	return nil
}

// 拨号配置
// WithDialOption returns a func to customize a ClientOptions with given dial option.
func WithDialOption(opt grpc.DialOption) ClientOption {
	return func(options *ClientOptions) {
		options.DialOptions = append(options.DialOptions, opt)
	}
}

// 非阻塞拨号设置
// WithNonBlock sets the dialing to be nonblock.
func WithNonBlock() ClientOption {
	return func(options *ClientOptions) {
		options.NonBlock = true
	}
}

// 超时设置
// WithTimeout returns a func to customize a ClientOptions with given timeout.
func WithTimeout(timeout time.Duration) ClientOption {
	return func(options *ClientOptions) {
		options.Timeout = timeout
	}
}

// Grpc调用凭据设置
// WithTransportCredentials return a func to make the gRPC calls secured with given credentials.
func WithTransportCredentials(creds credentials.TransportCredentials) ClientOption {
	return func(options *ClientOptions) {
		options.Secure = true
		options.DialOptions = append(options.DialOptions, grpc.WithTransportCredentials(creds))
	}
}

// 自定义拦截器设置
// WithUnaryClientInterceptor returns a func to customize a ClientOptions with given interceptor.
func WithUnaryClientInterceptor(interceptor grpc.UnaryClientInterceptor) ClientOption {
	return func(options *ClientOptions) {
		options.DialOptions = append(options.DialOptions, WithUnaryClientInterceptors(interceptor))
	}
}
