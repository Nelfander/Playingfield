package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/nelfander/Playingfield/internal/domain/user"
	"github.com/nelfander/Playingfield/internal/infrastructure/auth"
	"github.com/nelfander/Playingfield/internal/interfaces/http/dto"
	"github.com/nelfander/Playingfield/internal/interfaces/http/handlers"
	"github.com/stretchr/testify/assert"
)

// setupHandler returns both the UserHandler and the underlying FakeRepository
func setupHandler() (*handlers.UserHandler, *user.FakeRepository) {
	fakeRepo := user.NewFakeRepository()
	service := user.NewService(fakeRepo)
	jwtManager := auth.NewJWTManager("test-secret", 24*time.Hour)
	handler := handlers.NewUserHandler(service, jwtManager)
	return handler, fakeRepo
}

func TestUserRegistration(t *testing.T) {
	handler, _ := setupHandler()
	e := echo.New()

	reqBody := `{"email":"test@example.com","password":"supersecret"}`
	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if assert.NoError(t, handler.Register(c)) {
		var resp map[string]interface{}
		assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		assert.Equal(t, "user", resp["role"])
		assert.Equal(t, http.StatusCreated, rec.Code)
	}
}

func TestUserLogin(t *testing.T) {
	handler, fakeRepo := setupHandler()
	e := echo.New()

	// Prepare user manually in fake repo
	password := "supersecret"
	hashed, err := auth.HashPassword(password)
	if err != nil {
		t.Fatal(err)
	}

	fakeRepo.Users = append(fakeRepo.Users, user.User{
		Email:        "login@example.com",
		PasswordHash: hashed,
		Role:         "user",
		Status:       "active",
	})

	// Login request
	loginBody := `{"email":"login@example.com","password":"supersecret"}`
	reqLogin := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(loginBody))
	reqLogin.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	recLogin := httptest.NewRecorder()
	cLogin := e.NewContext(reqLogin, recLogin)

	if assert.NoError(t, handler.Login(cLogin)) {
		assert.Equal(t, http.StatusOK, recLogin.Code)
		var resp map[string]interface{}
		assert.NoError(t, json.Unmarshal(recLogin.Body.Bytes(), &resp))
		assert.NotEmpty(t, resp["token"])
	}
}

func TestUserLogin_InvalidCredentials(t *testing.T) {
	handler, _ := setupHandler()
	e := echo.New()

	reqBody := `{"email":"invalid@example.com","password":"wrong"}`
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler.Login(c)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	var resp map[string]interface{}
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "invalid credentials", resp["error"])
}

func TestUserLogin_InactiveAccount(t *testing.T) {
	handler, fakeRepo := setupHandler()
	e := echo.New()

	// Insert inactive user
	password := "secret"
	hashed, err := auth.HashPassword(password)
	if err != nil {
		t.Fatal(err)
	}

	fakeRepo.Users = append(fakeRepo.Users, user.User{
		Email:        "inactive@example.com",
		PasswordHash: hashed,
		Role:         "user",
		Status:       "inactive",
	})

	// Attempt login
	loginBody := `{"email":"inactive@example.com","password":"secret"}`
	reqLogin := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(loginBody))
	reqLogin.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	recLogin := httptest.NewRecorder()
	cLogin := e.NewContext(reqLogin, recLogin)

	handler.Login(cLogin)

	assert.Equal(t, http.StatusForbidden, recLogin.Code)
	var resp map[string]interface{}
	assert.NoError(t, json.Unmarshal(recLogin.Body.Bytes(), &resp))
	assert.Equal(t, "account is inactive or banned", resp["error"])
}

func TestMeEndpoint(t *testing.T) {
	//  Create fake repo and service locally
	fakeRepo := user.NewFakeRepository()
	service := user.NewService(fakeRepo)

	//  Register a user via the Service
	fakeUser, err := service.RegisterUser(context.Background(), "me@example.com", "supersecret")
	assert.NoError(t, err)

	//  Create JWT manager and handler
	jwtManager := auth.NewJWTManager("test-secret", 24*time.Hour)
	handler := handlers.NewUserHandler(service, jwtManager)

	//  Prepare echo request/recorder
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	//  Inject JWT claims manually
	claims := &auth.Claims{
		UserID: fakeUser.ID,
		Email:  fakeUser.Email,
		Role:   fakeUser.Role,
		Status: fakeUser.Status,
	}
	c.Set("user", claims)

	//  Call /me handler
	err = handler.Me(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp dto.UserResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "me@example.com", resp.Email)
	assert.Equal(t, "user", resp.Role)
	assert.Equal(t, "active", resp.Status)
}
