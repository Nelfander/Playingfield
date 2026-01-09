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
	auth    *auth.JWTManager // for future login
}

func NewUserHandler(s *user.Service, a *auth.JWTManager) *UserHandler {
	return &UserHandler{
		service: s,
		auth:    a,
	}
}

// Register handles POST /users
func (h *UserHandler) Register(c echo.Context) error {
	var req dto.RegisterUserRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request"})
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to hash password"})
	}

	u, err := h.service.RegisterUser(c.Request().Context(), req.Email, hash)
	if err != nil {
		if err == user.ErrUserAlreadyExists {
			return c.JSON(http.StatusConflict, echo.Map{"error": "user already exists"})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "internal error"})
	}

	// Respond with safe JSON
	resp := dto.UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
	}

	return c.JSON(http.StatusCreated, resp)
}

// Login handles POST /login
func (h *UserHandler) Login(c echo.Context) error {
	var req dto.LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request"})
	}

	// Call domain service
	u, err := h.service.Login(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "invalid credentials"})
	}

	// Generate JWT
	token, err := h.auth.GenerateToken(u.ID, u.Email, u.Role)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to generate token"})
	}

	// Map domain User -> DTO
	resp := dto.LoginResponse{
		Token: token,
		User: dto.UserResponse{
			ID:        u.ID,
			Email:     u.Email,
			Role:      u.Role,
			CreatedAt: u.CreatedAt,
		},
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *UserHandler) Me(c echo.Context) error {
	claims := c.Get("user").(*auth.Claims)
	return c.JSON(http.StatusOK, map[string]interface{}{
		"id":    claims.UserID,
		"email": claims.Email,
		"role":  claims.Role,
	})
}

func (h *UserHandler) Admin(c echo.Context) error {
	claims := c.Get("user").(*auth.Claims)
	return c.JSON(http.StatusOK, map[string]string{
		"message": "Welcome, admin " + claims.Email,
	})
}
