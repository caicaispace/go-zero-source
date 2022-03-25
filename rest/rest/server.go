package rest

import (
	"crypto/tls"
	"log"
	"net/http"
	"path"
	"time"

	"gozerosource/rest/rest/internal/cors"

	"gozerosource/rest/rest/handler"
	"gozerosource/rest/rest/httpx"
	"gozerosource/rest/rest/router"

	"github.com/zeromicro/go-zero/core/logx"
)

type (
	// RunOption defines the method to customize a Server.
	RunOption func(*Server)

	// A Server is a http server.
	Server struct {
		ngin   *engine
		router httpx.Router
	}
)

// MustNewServer returns a server with given config of c and options defined in opts.
// Be aware that later RunOption might overwrite previous one that write the same option.
// The process will exit if error occurs.
// 初始化（如有出错直接退出）
func MustNewServer(c RestConf, opts ...RunOption) *Server {
	server, err := NewServer(c, opts...)
	if err != nil {
		log.Fatal(err)
	}

	return server
}

// NewServer returns a server with given config of c and options defined in opts.
// Be aware that later RunOption might overwrite previous one that write the same option.
// 初始化
func NewServer(c RestConf, opts ...RunOption) (*Server, error) {
	if err := c.SetUp(); err != nil {
		return nil, err
	}

	server := &Server{
		ngin:   newEngine(c),       // 加载核心引擎
		router: router.NewRouter(), // 加载路由
	}

	opts = append([]RunOption{WithNotFoundHandler(nil)}, opts...) // 加载路由未找到方法
	for _, opt := range opts {
		opt(server) // 加载运行时方法
	}

	return server, nil
}

// AddRoutes add given routes into the Server.
// 批量添加路由
func (s *Server) AddRoutes(rs []Route, opts ...RouteOption) {
	r := featuredRoutes{
		routes: rs,
	}
	for _, opt := range opts {
		opt(&r)
	}
	s.ngin.addRoutes(r)
}

// AddRoute adds given route into the Server.
// 添加路由
func (s *Server) AddRoute(r Route, opts ...RouteOption) {
	s.AddRoutes([]Route{r}, opts...)
}

// Start starts the Server.
// Graceful shutdown is enabled by default.
// Use proc.SetTimeToForceQuit to customize the graceful shutdown period.
// 启动服务
func (s *Server) Start() {
	handleError(s.ngin.start(s.router))
}

// Stop stops the Server.
// 停止服务
func (s *Server) Stop() {
	logx.Close()
}

// Use adds the given middleware in the Server.
// 加载中间件
func (s *Server) Use(middleware Middleware) {
	s.ngin.use(middleware)
}

// ToMiddleware converts the given handler to a Middleware.
// 将 handle 转换为 Middleware
func ToMiddleware(handler func(next http.Handler) http.Handler) Middleware {
	return func(handle http.HandlerFunc) http.HandlerFunc {
		return handler(handle).ServeHTTP
	}
}

// WithCors returns a func to enable CORS for given origin, or default to all origins (*).
// 跨域处理器
func WithCors(origin ...string) RunOption {
	return func(server *Server) {
		server.router.SetNotAllowedHandler(cors.NotAllowedHandler(nil, origin...))
		server.Use(cors.Middleware(nil, origin...))
	}
}

// WithCustomCors returns a func to enable CORS for given origin, or default to all origins (*),
// fn lets caller customizing the response.
// 自定义跨域处理器
func WithCustomCors(middlewareFn func(header http.Header), notAllowedFn func(http.ResponseWriter),
	origin ...string,
) RunOption {
	return func(server *Server) {
		server.router.SetNotAllowedHandler(cors.NotAllowedHandler(notAllowedFn, origin...))
		server.Use(cors.Middleware(middlewareFn, origin...))
	}
}

// WithJwt returns a func to enable jwt authentication in given route.
// jwt 处理器
func WithJwt(secret string) RouteOption {
	return func(r *featuredRoutes) {
		validateSecret(secret)
		r.jwt.enabled = true
		r.jwt.secret = secret
	}
}

// WithJwtTransition returns a func to enable jwt authentication as well as jwt secret transition.
// Which means old and new jwt secrets work together for a period.
func WithJwtTransition(secret, prevSecret string) RouteOption {
	return func(r *featuredRoutes) {
		// why not validate prevSecret, because prevSecret is an already used one,
		// even it not meet our requirement, we still need to allow the transition.
		validateSecret(secret)
		r.jwt.enabled = true
		r.jwt.secret = secret
		r.jwt.prevSecret = prevSecret
	}
}

// WithMiddlewares adds given middlewares to given routes.
// jwt token 转换器，新老 token 可以同时使用
func WithMiddlewares(ms []Middleware, rs ...Route) []Route {
	for i := len(ms) - 1; i >= 0; i-- {
		rs = WithMiddleware(ms[i], rs...)
	}
	return rs
}

// WithMiddleware adds given middleware to given route.
// 给指定路由加载中间件
func WithMiddleware(middleware Middleware, rs ...Route) []Route {
	routes := make([]Route, len(rs))

	for i := range rs {
		route := rs[i]
		routes[i] = Route{
			Method:  route.Method,
			Path:    route.Path,
			Handler: middleware(route.Handler),
		}
	}

	return routes
}

// WithNotFoundHandler returns a RunOption with not found handler set to given handler.
// 路由未找到处理方法
func WithNotFoundHandler(handler http.Handler) RunOption {
	return func(server *Server) {
		notFoundHandler := server.ngin.notFoundHandler(handler)
		server.router.SetNotFoundHandler(notFoundHandler)
	}
}

// WithNotAllowedHandler returns a RunOption with not allowed handler set to given handler.
// 不予通过处理方法
func WithNotAllowedHandler(handler http.Handler) RunOption {
	return func(server *Server) {
		server.router.SetNotAllowedHandler(handler)
	}
}

// WithPrefix adds group as a prefix to the route paths.
// 路由前缀处理方法
func WithPrefix(group string) RouteOption {
	return func(r *featuredRoutes) {
		var routes []Route
		for _, rt := range r.routes {
			p := path.Join(group, rt.Path)
			routes = append(routes, Route{
				Method:  rt.Method,
				Path:    p,
				Handler: rt.Handler,
			})
		}
		r.routes = routes
	}
}

// WithPriority returns a RunOption with priority.
// 给路 featuredRoutes 提高优先级，在熔断服务中使 featuredRoutes 有更高的熔断阀值
func WithPriority() RouteOption {
	return func(r *featuredRoutes) {
		r.priority = true
	}
}

// WithRouter returns a RunOption that make server run with given router.
// server 使用给定路由
func WithRouter(router httpx.Router) RunOption {
	return func(server *Server) {
		server.router = router
	}
}

// WithSignature returns a RouteOption to enable signature verification.
// server 配置并开启签名验证
func WithSignature(signature SignatureConf) RouteOption {
	return func(r *featuredRoutes) {
		r.signature.enabled = true
		r.signature.Strict = signature.Strict
		r.signature.Expiry = signature.Expiry
		r.signature.PrivateKeys = signature.PrivateKeys
	}
}

// WithTimeout returns a RouteOption to set timeout with given value.
// server 配置并开启服务超时
func WithTimeout(timeout time.Duration) RouteOption {
	return func(r *featuredRoutes) {
		r.timeout = timeout
	}
}

// WithTLSConfig returns a RunOption that with given tls config.
// server 配置并开启 TLS
func WithTLSConfig(cfg *tls.Config) RunOption {
	return func(srv *Server) {
		srv.ngin.setTlsConfig(cfg)
	}
}

// WithUnauthorizedCallback returns a RunOption that with given unauthorized callback set.
// server 配置未授权回调函数
func WithUnauthorizedCallback(callback handler.UnauthorizedCallback) RunOption {
	return func(srv *Server) {
		srv.ngin.setUnauthorizedCallback(callback)
	}
}

// WithUnsignedCallback returns a RunOption that with given unsigned callback set.
// server 配置未签名回调函数
func WithUnsignedCallback(callback handler.UnsignedCallback) RunOption {
	return func(srv *Server) {
		srv.ngin.setUnsignedCallback(callback)
	}
}

// 记录并抛出错误
func handleError(err error) {
	// ErrServerClosed means the server is closed manually
	if err == nil || err == http.ErrServerClosed {
		return
	}

	logx.Error(err)
	panic(err)
}

// secret 字符规范验证
func validateSecret(secret string) {
	if len(secret) < 8 {
		panic("secret's length can't be less than 8")
	}
}
