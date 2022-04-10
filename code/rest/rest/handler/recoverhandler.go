package handler

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"gozerosource/code/rest/rest/internal"
)

// RecoverHandler returns a middleware that recovers if panic happens.
// 错误捕获中间件
func RecoverHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if result := recover(); result != nil {
				internal.Error(r, fmt.Sprintf("%v\n%s", result, debug.Stack()))
				w.WriteHeader(http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
