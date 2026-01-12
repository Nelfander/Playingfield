package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/nelfander/Playingfield/internal/infrastructure/auth"
)

// JWTMiddleware verifies a JWT token and stores claims in context
func JWTMiddleware(jwtManager *auth.JWTManager) echo.MiddlewareFunc {
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

			// Store full claims
			c.Set("user", claims)

			c.Set("user_id", claims.UserID)

			return next(c)
		}
	}
}
