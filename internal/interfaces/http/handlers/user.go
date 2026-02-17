package handlers

import (
	"context"
	"net/http"
	"strconv"

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
		return err // The Translator will give the user a 400 with a detailed reason
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		return err // If hashing fails, it's a 500 error(translator handles it)
	}

	u, err := h.service.RegisterUser(c.Request().Context(), req.Email, hash)
	if err != nil {
		return err // translator maps ErrUserAlreadyExists to 409 Conflict
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
		return err
	}

	// call domain service
	u, err := h.service.Login(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		return err // the translator will decide if this is a 401 or 403
	}

	// generate JWT
	token, err := h.auth.GenerateToken(u.ID, u.Email, u.Role)
	if err != nil {
		return err
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
		return echo.NewHTTPError(http.StatusUnauthorized, "missing or invalid token")
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
		return err
	}

	return c.JSON(http.StatusOK, users)
}

// DELETE /admin/users/:id
func (h *UserHandler) ScrubUser(c echo.Context) error {
	idParam := c.Param("id")
	userID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user id")
	}

	// Call service layer
	if err := h.service.AdminScrubUser(c.Request().Context(), userID); err != nil {
		return err // translator will handle mapping this to a 500
	}

	return c.NoContent(http.StatusNoContent)
}

// GET /admin/users
func (h *UserHandler) AdminListAllUsers(c echo.Context) error {
	users, err := h.service.ListAllUsers(c.Request().Context())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, users)
}
