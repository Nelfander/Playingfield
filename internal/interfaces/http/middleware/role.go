package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/nelfander/Playingfield/internal/infrastructure/auth"
)

func RequireRole(jwtManager *auth.JWTManager, allowedRoles ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{"message": "missing authorization header"})
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				return c.JSON(http.StatusUnauthorized, map[string]string{"message": "invalid authorization header"})
			}

			claims, err := jwtManager.VerifyToken(parts[1])
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{"message": "invalid or expired token"})
			}

			// Check if user's role is allowed
			allowed := false
			for _, role := range allowedRoles {
				if claims.Role == role {
					allowed = true
					break
				}
			}

			if !allowed {
				return c.JSON(http.StatusForbidden, map[string]string{"message": "forbidden: insufficient role"})
			}

			// Store claims in context for handler use
			c.Set("user", claims)

			return next(c)
		}
	}
}
