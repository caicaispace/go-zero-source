package internal

import "google.golang.org/grpc"

// 服务端流拦截器
// WithStreamServerInterceptors uses given server stream interceptors.
func WithStreamServerInterceptors(interceptors ...grpc.StreamServerInterceptor) grpc.ServerOption {
	return grpc.ChainStreamInterceptor(interceptors...)
}

// 服务端拦截器
// WithUnaryServerInterceptors uses given server unary interceptors.
func WithUnaryServerInterceptors(interceptors ...grpc.UnaryServerInterceptor) grpc.ServerOption {
	return grpc.ChainUnaryInterceptor(interceptors...)
}
