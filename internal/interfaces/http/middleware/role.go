package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/nelfander/Playingfield/internal/infrastructure/auth"
)

// RequireRole ensures the authenticated user has at least one of the allowed roles.
// It relies on JWTMiddleware having already set valid claims in the context.
func RequireRole(_ *auth.JWTManager, allowedRoles ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			claims, ok := c.Get("user").(*auth.Claims)
			if !ok || claims == nil {
				// This should almost never happen if JWTMiddleware ran before
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"message": "authentication required",
				})
			}

			for _, allowed := range allowedRoles {
				if claims.Role == allowed {
					return next(c)
				}
			}

			return c.JSON(http.StatusForbidden, map[string]string{
				"message": "forbidden: insufficient permissions",
			})
		}
	}
}
