package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/nelfander/Playingfield/internal/infrastructure/auth"
	custom "github.com/nelfander/Playingfield/internal/interfaces/http"
	"github.com/nelfander/Playingfield/internal/interfaces/http/middleware"
	"github.com/stretchr/testify/assert"
)

func TestRateLimit_Integration(t *testing.T) {
	// setup environment
	e := echo.New()
	e.HTTPErrorHandler = custom.CustomHTTPErrorHandler
	jwtManager := auth.NewJWTManager("test_secret_32_characters_long_!!", time.Hour)

	e.Use(middleware.RateLimitMiddleware(jwtManager))

	e.GET("/ping", func(c echo.Context) error {
		return c.String(http.StatusOK, "pong")
	})

	t.Run("Anonymous IP Throttling", func(t *testing.T) {
		middleware.ResetVisitors()

		targetIP := "192.168.1.1"
		limitTriggered := false

		// Burst is 10 for anonymous. 11th or 12th should trigger err 429.
		for i := 0; i < 12; i++ {
			req := httptest.NewRequest(http.MethodGet, "/ping", nil)
			req.RemoteAddr = targetIP + ":1234"
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			if rec.Code == http.StatusTooManyRequests {
				limitTriggered = true
				assert.Contains(t, rec.Body.String(), "too many requests")
				break
			}
		}
		assert.True(t, limitTriggered, "Rate limit should have been triggered for anonymous user")
	})

	t.Run("Authenticated User Limits", func(t *testing.T) {
		middleware.ResetVisitors()

		// Generate a real token for User 999
		token, err := jwtManager.GenerateToken(999, "auth@example.com", "user")
		assert.NoError(t, err)

		// auth burst is 40.
		// if we send 25 requests, it SHOULD still be 200 OK.
		// (if it was anonymous it would have failed at 11).
		for i := 0; i < 25; i++ {
			req := httptest.NewRequest(http.MethodGet, "/ping", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code, "Authenticated user should not be limited at 25 requests")
		}
	})
}
