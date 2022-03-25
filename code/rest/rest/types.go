package rest

import (
	"net/http"
	"time"
)

type (
	// Middleware defines the middleware method.
	// 中间件标准类型
	Middleware func(next http.HandlerFunc) http.HandlerFunc

	// A Route is a http route.
	// 路由
	Route struct {
		Method  string           // http 方法
		Path    string           // 路由 path
		Handler http.HandlerFunc // 处理函数
	}

	// RouteOption defines the method to customize a featured route.
	// 定义特色路由
	RouteOption func(r *featuredRoutes)

	// jwt
	jwtSetting struct {
		enabled    bool   // 是否开启
		secret     string // 加密串
		prevSecret string // 前一个加密串（兼容处理）
	}

	// 签名
	signatureSetting struct {
		SignatureConf      // 签名配置
		enabled       bool // 是否开启
	}

	// 特色路由
	featuredRoutes struct {
		timeout   time.Duration    // 超时处理
		priority  bool             // 是否开启优先
		jwt       jwtSetting       // jwt 配置
		signature signatureSetting // 签名配置
		routes    []Route          // 指定路由
	}
)
