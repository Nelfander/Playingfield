package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/nelfander/Playingfield/internal/infrastructure/auth"
	"github.com/nelfander/Playingfield/internal/interfaces/http/middleware"
	"github.com/stretchr/testify/assert"
)

func TestJWTMiddleware_MissingAuthorizationHeader(t *testing.T) {
	e := echo.New()

	jwtManager := auth.NewJWTManager("test-secret", time.Hour)

	// Dummy protected handler
	protectedHandler := func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"message": "ok"})
	}

	// Wrap handler with middleware
	handler := middleware.JWTMiddleware(jwtManager)(protectedHandler)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestJWTMiddleware_InvalidToken(t *testing.T) {
	e := echo.New()
	jwtManager := auth.NewJWTManager("test-secret", time.Hour)

	protectedHandler := func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"message": "ok"})
	}

	handler := middleware.JWTMiddleware(jwtManager)(protectedHandler)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

/*  Test valid token case
Proves that : Token verification works,
middleware does not swallow the request,
claims are injected into context,
handler executes exactly once
*/

func TestJWTMiddleware_ValidToken(t *testing.T) {
	e := echo.New()
	jwtManager := auth.NewJWTManager("test-secret", time.Hour)

	// Create a real token
	token, err := jwtManager.GenerateToken(1, "test@example.com", "user")
	assert.NoError(t, err)

	called := false

	protectedHandler := func(c echo.Context) error {
		called = true

		claims := c.Get("user")
		assert.NotNil(t, claims)

		return c.JSON(http.StatusOK, map[string]string{"message": "ok"})
	}

	handler := middleware.JWTMiddleware(jwtManager)(protectedHandler)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = handler(c)

	assert.NoError(t, err)
	assert.True(t, called)
	assert.Equal(t, http.StatusOK, rec.Code)
}
