package server

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v5"
)

// apiKeyAuth returns middleware that validates the Authorization header.
func apiKeyAuth(apiKey string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if c.Request().URL.Path == "/health" {
				return next(c)
			}

			auth := c.Request().Header.Get("Authorization")
			if auth == "" {
				return c.JSON(http.StatusUnauthorized, ErrorResponse{
					Error: ErrorDetail{
						Message: "Missing Authorization header",
						Type:    "authentication_error",
					},
				})
			}

			token := strings.TrimPrefix(auth, "Bearer ")
			if token != apiKey {
				return c.JSON(http.StatusUnauthorized, ErrorResponse{
					Error: ErrorDetail{
						Message: "Invalid API key",
						Type:    "authentication_error",
					},
				})
			}

			return next(c)
		}
	}
}
