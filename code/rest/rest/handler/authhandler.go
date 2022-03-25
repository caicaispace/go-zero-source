package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httputil"

	"gozerosource/code/rest/rest/internal/response"
	"gozerosource/code/rest/rest/token"

	"github.com/golang-jwt/jwt/v4"
	"github.com/zeromicro/go-zero/core/logx"
)

const (
	jwtAudience    = "aud"
	jwtExpire      = "exp"
	jwtId          = "jti"
	jwtIssueAt     = "iat"
	jwtIssuer      = "iss"
	jwtNotBefore   = "nbf"
	jwtSubject     = "sub"
	noDetailReason = "no detail reason"
)

var (
	errInvalidToken = errors.New("invalid auth token")
	errNoClaims     = errors.New("no auth params")
)

type (
	// A AuthorizeOptions is authorize options.
	AuthorizeOptions struct {
		PrevSecret string               // 上一个 Secret
		Callback   UnauthorizedCallback // 验证失败回调
	}

	// UnauthorizedCallback defines the method of unauthorized callback.
	// 验证失败标准函数
	UnauthorizedCallback func(w http.ResponseWriter, r *http.Request, err error)
	// AuthorizeOption defines the method to customize an AuthorizeOptions.
	// 权限验证标准配置
	AuthorizeOption func(opts *AuthorizeOptions)
)

// Authorize returns an authorize middleware.
// 权限验证中间件
func Authorize(secret string, opts ...AuthorizeOption) func(http.Handler) http.Handler {
	var authOpts AuthorizeOptions
	for _, opt := range opts {
		opt(&authOpts)
	}

	parser := token.NewTokenParser() // 加载 token 解析器
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 解析 token
			tok, err := parser.ParseToken(r, secret, authOpts.PrevSecret)
			if err != nil {
				unauthorized(w, r, err, authOpts.Callback)
				return
			}

			if !tok.Valid {
				unauthorized(w, r, errInvalidToken, authOpts.Callback)
				return
			}

			claims, ok := tok.Claims.(jwt.MapClaims)
			if !ok {
				unauthorized(w, r, errNoClaims, authOpts.Callback)
				return
			}

			ctx := r.Context() // 获取上下文
			for k, v := range claims {
				switch k {
				case jwtAudience, jwtExpire, jwtId, jwtIssueAt, jwtIssuer, jwtNotBefore, jwtSubject:
					// ignore the standard claims
					// 忽略 jwt 标准声明
				default:
					ctx = context.WithValue(ctx, k, v) // 解析后的数据注入上下文
				}
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// WithPrevSecret returns an AuthorizeOption with setting previous secret.
// 设置上一个secret
func WithPrevSecret(secret string) AuthorizeOption {
	return func(opts *AuthorizeOptions) {
		opts.PrevSecret = secret
	}
}

// WithUnauthorizedCallback returns an AuthorizeOption with setting unauthorized callback.
// 设置验证失败回调
func WithUnauthorizedCallback(callback UnauthorizedCallback) AuthorizeOption {
	return func(opts *AuthorizeOptions) {
		opts.Callback = callback
	}
}

// 记录详细日志
func detailAuthLog(r *http.Request, reason string) {
	// discard dump error, only for debug purpose
	details, _ := httputil.DumpRequest(r, true)
	logx.Errorf("authorize failed: %s\n=> %+v", reason, string(details))
}

// 加载验证失败回调&记录日志
func unauthorized(w http.ResponseWriter, r *http.Request, err error, callback UnauthorizedCallback) {
	writer := response.NewHeaderOnceResponseWriter(w)

	if err != nil {
		detailAuthLog(r, err.Error())
	} else {
		detailAuthLog(r, noDetailReason)
	}

	// let callback go first, to make sure we respond with user-defined HTTP header
	if callback != nil {
		callback(writer, r, err)
	}

	// if user not setting HTTP header, we set header with 401
	writer.WriteHeader(http.StatusUnauthorized)
}
