package middleware

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"net/http"
)

// HTTPHandlerTracerMiddleware tracer for http server handler
func HTTPHandlerTracerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tracer := otel.GetTracerProvider().Tracer(instrumentationName)
		ctx, span := tracer.Start(r.Context(), "http-handler-tracer")
		span.SetAttributes(attribute.String("http.method", r.Method))
		span.SetAttributes(attribute.String("http.url", r.URL.String()))
		defer span.End()
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
