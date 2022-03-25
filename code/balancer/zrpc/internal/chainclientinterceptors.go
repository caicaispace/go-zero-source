package internal

import "google.golang.org/grpc"

// 客户端流拦截器
// WithStreamClientInterceptors uses given client stream interceptors.
func WithStreamClientInterceptors(interceptors ...grpc.StreamClientInterceptor) grpc.DialOption {
	return grpc.WithChainStreamInterceptor(interceptors...)
}

// 客户端拦截器
// WithUnaryClientInterceptors uses given client unary interceptors.
func WithUnaryClientInterceptors(interceptors ...grpc.UnaryClientInterceptor) grpc.DialOption {
	return grpc.WithChainUnaryInterceptor(interceptors...)
}
