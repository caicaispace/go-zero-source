package serverinterceptors

import (
	"context"

	"gozerosource/balancer/zrpc/internal/auth"

	"google.golang.org/grpc"
)

// 权限拦截器（数据流）
// StreamAuthorizeInterceptor returns a func that uses given authenticator in processing stream requests.
func StreamAuthorizeInterceptor(authenticator *auth.Authenticator) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		if err := authenticator.Authenticate(stream.Context()); err != nil {
			return err
		}

		return handler(srv, stream)
	}
}

// 权限拦截器
// UnaryAuthorizeInterceptor returns a func that uses given authenticator in processing unary requests.
func UnaryAuthorizeInterceptor(authenticator *auth.Authenticator) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if err := authenticator.Authenticate(ctx); err != nil {
			return nil, err
		}

		return handler(ctx, req)
	}
}
