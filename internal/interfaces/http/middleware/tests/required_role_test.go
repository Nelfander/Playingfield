package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

	"github.com/nelfander/Playingfield/internal/infrastructure/auth"
	"github.com/nelfander/Playingfield/internal/interfaces/http/middleware"
)

func TestRequireRole_AdminAllowed(t *testing.T) {
	e := echo.New()

	jwtManager := auth.NewJWTManager("test-secret", time.Hour)
	token, _ := jwtManager.GenerateToken(1, "admin@test.com", "admin")

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	handler := middleware.RequireRole(jwtManager, "admin")(func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	err := handler(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRequireRole_UserForbidden(t *testing.T) {
	e := echo.New()

	jwtManager := auth.NewJWTManager("test-secret", time.Hour)
	token, _ := jwtManager.GenerateToken(1, "user@test.com", "user")

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	handler := middleware.RequireRole(jwtManager, "admin")(func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	err := handler(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestRequireRole_MissingToken(t *testing.T) {
	e := echo.New()

	jwtManager := auth.NewJWTManager("test-secret", time.Hour)

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	handler := middleware.RequireRole(jwtManager, "admin")(func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	err := handler(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
