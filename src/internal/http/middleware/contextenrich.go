package middleware

import (
	"github.com/labstack/echo/v4"
)

const EnrichReqIDKey string = "enrich:request_id"

func NewEnrichRequestContextMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set(EnrichReqIDKey, c.Response().Header().Get(echo.HeaderXRequestID))
			err := next(c)
			return err
		}
	}
}
