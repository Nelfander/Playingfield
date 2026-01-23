package handlers

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/nelfander/Playingfield/internal/domain/user"
	"github.com/nelfander/Playingfield/internal/infrastructure/auth"
	"github.com/nelfander/Playingfield/internal/interfaces/http/dto"
)

type UserHandler struct {
	service user.Service
	auth    *auth.JWTManager
}

// for test purposes
func (h *UserHandler) RegisterUserForTest(email, password string) (*user.User, error) {
	return h.service.RegisterUser(context.Background(), email, password)
}

// generate JWT token directly via auth (for testing)
func (h *UserHandler) GenerateTokenForTest(id int64, email, role string) (string, error) {
	return h.auth.GenerateToken(id, email, role)
}

func NewUserHandler(service user.Service, auth *auth.JWTManager) *UserHandler {
	return &UserHandler{service: service, auth: auth}
}

// register handles POST /users
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

	resp := dto.UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		Role:      u.Role,
		Status:    u.Status,
		CreatedAt: u.CreatedAt,
	}

	return c.JSON(http.StatusCreated, resp)
}

// login handles POST /login
func (h *UserHandler) Login(c echo.Context) error {
	var req dto.LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request"})
	}

	// call domain service
	u, err := h.service.Login(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		if err == user.ErrInactiveAccount {
			return c.JSON(http.StatusForbidden, echo.Map{"error": err.Error()})
		}
		// all the other errors (wrong credentials, etc.)
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "invalid credentials"})
	}

	// generate JWT
	token, err := h.auth.GenerateToken(u.ID, u.Email, u.Role)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to generate token"})
	}

	// map domain User -> DTO
	resp := dto.LoginResponse{
		Token:  token,
		UserId: u.ID,
		User: dto.UserResponse{
			ID:        u.ID,
			Email:     u.Email,
			Role:      u.Role,
			Status:    u.Status,
			CreatedAt: u.CreatedAt,
		},
	}

	return c.JSON(http.StatusOK, resp)
}

// Me handles GET /me
func (h *UserHandler) Me(c echo.Context) error {
	// grab claims from context (set by JWT middleware)
	claims, ok := c.Get("user").(*auth.Claims)
	if !ok || claims == nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "unauthorized"})
	}

	resp := dto.UserResponse{
		ID:     claims.UserID,
		Email:  claims.Email,
		Role:   claims.Role,
		Status: claims.Status,
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *UserHandler) Admin(c echo.Context) error {
	claims := c.Get("user").(*auth.Claims)
	return c.JSON(http.StatusOK, map[string]string{
		"message": "Welcome, admin " + claims.Email,
	})
}

// GET /users
func (h *UserHandler) List(c echo.Context) error {
	users, err := h.service.ListAllUsers(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to fetch users"})
	}

	return c.JSON(http.StatusOK, users)
}
