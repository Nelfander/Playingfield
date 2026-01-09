package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/nelfander/Playingfield/internal/infrastructure/auth"
)

// RequireRole checks if the user has one of the allowed roles
func RequireRole(jwtManager *auth.JWTManager, allowedRoles ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") {
				return c.JSON(http.StatusUnauthorized, map[string]string{"message": "invalid or missing authorization header"})
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			claims, err := jwtManager.VerifyToken(tokenStr)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{"message": "invalid or expired token"})
			}

			for _, role := range allowedRoles {
				if claims.Role == role {
					c.Set("user", claims) // store claims for handler
					return next(c)
				}
			}

			return c.JSON(http.StatusForbidden, map[string]string{"message": "forbidden: insufficient role"})
		}
	}
}
