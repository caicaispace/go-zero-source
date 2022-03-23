package clientinterceptors

import (
	"context"
	"path"

	"gozerosource/balancer/zrpc/internal/codes"

	"github.com/zeromicro/go-zero/core/breaker"

	"google.golang.org/grpc"
)

// 断路拦截器
// BreakerInterceptor is an interceptor that acts as a circuit breaker.
func BreakerInterceptor(ctx context.Context, method string, req, reply interface{},
	cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption,
) error {
	breakerName := path.Join(cc.Target(), method)
	return breaker.DoWithAcceptable(breakerName, func() error {
		return invoker(ctx, method, req, reply, cc, opts...)
	}, codes.Acceptable)
}
