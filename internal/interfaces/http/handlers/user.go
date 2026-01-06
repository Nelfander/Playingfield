package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/nelfander/Playingfield/internal/domain/user"
	"github.com/nelfander/Playingfield/internal/infrastructure/auth"
	"github.com/nelfander/Playingfield/internal/interfaces/http/dto"
)

type UserHandler struct {
	service *user.Service
}

func NewUserHandler(service *user.Service) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) Register(c echo.Context) error {
	var req dto.RegisterUserRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error": "invalid request",
		})
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error": "failed to process password",
		})
	}

	_, err = h.service.RegisterUser(
		c.Request().Context(),
		req.Email,
		hash,
	)

	if err != nil {
		switch err {
		case user.ErrUserAlreadyExists:
			return c.JSON(http.StatusConflict, echo.Map{
				"error": "user already exists",
			})
		default:
			return c.JSON(http.StatusInternalServerError, echo.Map{
				"error": "internal error",
			})
		}
	}

	return c.JSON(http.StatusCreated, echo.Map{
		"status": "ok",
	})
}
