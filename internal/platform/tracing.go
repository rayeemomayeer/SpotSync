package platform

import (
	"os"

	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
)

// TracingMiddleware returns Echo OTel middleware when tracing is enabled.
func TracingMiddleware(serviceName string) echo.MiddlewareFunc {
	if os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") == "" {
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			return next
		}
	}
	return otelecho.Middleware(serviceName)
}
