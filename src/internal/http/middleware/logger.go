package middleware

import (
	"log/slog"
	"time"

	"github.com/labstack/echo/v4"
)

func NewRequestLoggerMiddleware(log *slog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		log := log.With(
			slog.String("component", "middleware.logger"),
		)

		log.Info("logger middleware enabled")

		return func(c echo.Context) error {
			req := c.Request()
			entry := log.With(
				slog.String("method", req.Method),
				slog.String("path", req.URL.Path),
				slog.String("remote_addr", req.RemoteAddr),
				slog.String("user_agent", req.UserAgent()),
				slog.String("request_id", c.Response().Header().Get(echo.HeaderXRequestID)),
			)

			t1 := time.Now()
			err := next(c)
			entry.Info("request completed",
				slog.Int("status", c.Response().Status),
				slog.Int64("bytes", c.Response().Size),
				slog.String("duration", time.Since(t1).String()),
			)

			return err
		}
	}
}
