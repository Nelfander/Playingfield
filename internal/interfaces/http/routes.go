package http

import (
	"github.com/labstack/echo/v4"
	"github.com/nelfander/Playingfield/internal/interfaces/http/handlers"
)

func RegisterRoutes(e *echo.Echo, userHandler *handlers.UserHandler) {
	// Health route
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	})

	// User routes
	e.POST("/users", userHandler.Register)
	e.POST("/login", userHandler.Login)
}
