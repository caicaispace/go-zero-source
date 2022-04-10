package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/core/trace"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	oteltrace "go.opentelemetry.io/otel/trace"
)

// TracingHandler return a middleware that process the opentelemetry.
// 链路追踪中间件
func TracingHandler(serviceName, path string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		// GetTextMapPropagator返回全局TextMapPropagator。如果没有
		// 设置，将返回一个No-Op TextMapPropagator。
		propagator := otel.GetTextMapPropagator()
		// GetTracerProvider返回已注册的全局跟踪器。
		// 如果没有注册，则返回NoopTracerProvider的实例。
		// 使用跟踪提供者来创建一个命名的跟踪器。例如。
		// tracer := otel.GetTracerProvider().Tracer("example.com/foo")
		// 或
		// tracer := otel.Tracer("example.com/foo")
		tracer := otel.GetTracerProvider().Tracer(trace.TraceName)

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := propagator.Extract(r.Context(), propagation.HeaderCarrier(r.Header))
			spanName := path
			if len(spanName) == 0 {
				spanName = r.URL.Path
			}
			spanCtx, span := tracer.Start(
				ctx,
				spanName,
				oteltrace.WithSpanKind(oteltrace.SpanKindServer),
				oteltrace.WithAttributes(semconv.HTTPServerAttributesFromHTTPRequest(
					serviceName, spanName, r)...),
			)
			defer span.End()

			// convenient for tracking error messages
			sc := span.SpanContext()
			if sc.HasTraceID() {
				w.Header().Set(trace.TraceIdKey, sc.TraceID().String())
			}

			next.ServeHTTP(w, r.WithContext(spanCtx))
		})
	}
}
