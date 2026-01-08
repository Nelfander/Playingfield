package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func Me(c echo.Context) error {
	return c.JSON(http.StatusOK, echo.Map{
		"id":    c.Get("user_id"),
		"email": c.Get("email"),
		"role":  c.Get("role"),
	})
}
