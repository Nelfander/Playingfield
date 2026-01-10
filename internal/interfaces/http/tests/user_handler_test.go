package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/nelfander/Playingfield/internal/domain/user"
	"github.com/nelfander/Playingfield/internal/infrastructure/auth"
	"github.com/nelfander/Playingfield/internal/interfaces/http/handlers"
	"github.com/stretchr/testify/assert"
)

func setupHandler() *handlers.UserHandler {
	// Fake repo instead of real DB
	fakeRepo := user.NewFakeRepository()

	// Service with fake repo
	service := user.NewService(fakeRepo)

	// JWT manager
	jwtManager := auth.NewJWTManager("test-secret", 24*time.Hour)

	// Handler
	handler := handlers.NewUserHandler(service, jwtManager)
	return handler
}

func TestUserRegistration(t *testing.T) {
	handler := setupHandler()
	e := echo.New()

	// Prepare request body
	reqBody := `{"email":"test@example.com","password":"supersecret"}`
	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Call Register handler
	if assert.NoError(t, handler.Register(c)) {
		var resp map[string]interface{}
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		assert.NoError(t, err)

		// Check role is set correctly
		assert.Equal(t, "user", resp["role"])
		assert.Equal(t, http.StatusCreated, rec.Code)
	}
}

func TestUserLogin(t *testing.T) {
	handler := setupHandler()
	e := echo.New()

	// First, register the user
	registerBody := `{"email":"login@example.com","password":"supersecret"}`
	reqReg := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(registerBody))
	reqReg.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	recReg := httptest.NewRecorder()
	cReg := e.NewContext(reqReg, recReg)
	handler.Register(cReg)

	// Now login
	loginBody := `{"email":"login@example.com","password":"supersecret"}`
	reqLogin := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(loginBody))
	reqLogin.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	recLogin := httptest.NewRecorder()
	cLogin := e.NewContext(reqLogin, recLogin)

	if assert.NoError(t, handler.Login(cLogin)) {
		assert.Equal(t, http.StatusOK, recLogin.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(recLogin.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.NotEmpty(t, resp["token"])
	}
}

func TestUserLogin_InvalidCredentials(t *testing.T) {
	handler := setupHandler()
	e := echo.New()

	// Try login without registering
	loginBody := `{"email":"invalid@example.com","password":"wrong"}`
	reqLogin := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(loginBody))
	reqLogin.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	recLogin := httptest.NewRecorder()
	cLogin := e.NewContext(reqLogin, recLogin)

	handler.Login(cLogin) // no need to check returned error

	assert.Equal(t, http.StatusUnauthorized, recLogin.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(recLogin.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "invalid credentials", resp["error"])
}
