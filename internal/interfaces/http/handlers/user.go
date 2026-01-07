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
	auth    *auth.JWTManager
}
userHandler := handlers.NewUserHandler(userService, jwtManager)
func NewUserHandler(s *user.Service, a *auth.JWTManager) *UserHandler {
	return &UserHandler{
		service: s,
		auth:    a,
	}
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

	createdUser, err := h.service.RegisterUser(
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

	resp := dto.UserResponse{
		ID:        createdUser.ID,
		Email:     createdUser.Email,
		CreatedAt: createdUser.CreatedAt,
	}

	return c.JSON(http.StatusCreated, resp)
}

func (h *UserHandler) Login(c echo.Context) error {
	var req dto.LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request"})
	}

	user, err := h.service.Login(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "invalid credentials"})
	}

	token, err := h.auth.GenerateJWT(user.ID, user.Role)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to generate token"})
	}

	resp := dto.LoginResponse{
		Token: token,
		User: dto.UserResponse{
			ID:        user.ID,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
		},
	}

	return c.JSON(http.StatusOK, resp)
}
