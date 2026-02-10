package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestRateLimitMiddleware(t *testing.T) {
	// setup echo
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// use nil for jwtManager because we are testing the anonymous IP path
	handler := RateLimitMiddleware(nil)(func(c echo.Context) error {
		return c.String(http.StatusOK, "passed")
	})

	// test: initial requests should pass (Burst is 10)
	for i := 0; i < 10; i++ {
		err := handler(c)
		assert.NoError(t, err)
	}

	// test: The 11th request should be rate limited
	err := handler(c)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrRateLimitExceeded), "should return rate limit exceeded error")
}
