package middleware

import (
	"net/url"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	log "github.com/sirupsen/logrus"
)

// HTTPRequestLoggerWithConfig returns an echo.MiddlewareFunc with default logging config.
func HTTPRequestLoggerWithConfig() echo.MiddlewareFunc {
	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogRequestID:     true,
		LogURI:           true,
		LogStatus:        true,
		LogMethod:        true,
		LogLatency:       true,
		LogUserAgent:     true,
		LogProtocol:      true,
		LogResponseSize:  true,
		LogContentLength: true,
		LogRemoteIP:      true,
		LogHost:          true,
		LogError:         true,
		LogValuesFunc: func(_ echo.Context, v middleware.RequestLoggerValues) error {
			log.WithFields(log.Fields{
				"request_id":     v.RequestID,
				"uri":            v.URI,
				"unescaped_uri":  unescapeURI(v.URI),
				"status":         v.Status,
				"method":         v.Method,
				"duration_in_ms": v.Latency.Milliseconds(),
				"user_agent":     v.UserAgent,
				"protocol":       v.Protocol,
				"bytes_out":      v.ResponseSize,
				"bytes_in":       v.ContentLength,
				"host":           v.Host,
				"remote_ip":      v.RemoteIP,
				"error":          v.Error,
			}).Info("REQUEST INFO")
			return nil
		},
	})
}

func unescapeURI(in string) string {
	res, err := url.QueryUnescape(in)
	if err != nil {
		return in
	}
	return res
}
