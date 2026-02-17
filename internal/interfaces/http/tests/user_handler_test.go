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
	custom "github.com/nelfander/Playingfield/internal/interfaces/http"
	"github.com/nelfander/Playingfield/internal/interfaces/http/dto"
	"github.com/nelfander/Playingfield/internal/interfaces/http/handlers"
	"github.com/stretchr/testify/assert"
)

// setupHandler returns both the UserHandler and the underlying fakerepo
// clean way to get empty fakerepo for every test so that one test doesnt affect another
func setupHandler() (*handlers.UserHandler, *user.FakeRepository, *echo.Echo) {
	e := echo.New()
	e.HTTPErrorHandler = custom.CustomHTTPErrorHandler

	fakeRepo := user.NewFakeRepository()
	service := user.NewService(fakeRepo)
	jwtManager := auth.NewJWTManager("test-secret", 24*time.Hour)
	handler := handlers.NewUserHandler(service, jwtManager)
	return handler, fakeRepo, e
}

func TestUserRegistration(t *testing.T) {
	handler, _, e := setupHandler()

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
	handler, fakeRepo, e := setupHandler()

	// hash password manually to append in fake repo
	password := "supersecret"
	hashed, err := auth.HashPassword(password)
	if err != nil {
		t.Fatal(err)
	}

	// append user to fake repo
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
	handler, _, e := setupHandler()

	reqBody := `{"email":"invalid@example.com","password":"wrong"}`
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Capture the error and pass it to Echo's error handler
	err := handler.Login(c)
	if err != nil {
		e.HTTPErrorHandler(err, c)
	}

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	var resp map[string]interface{}
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "Invalid email or password", resp["error"])
}

func TestUserLogin_InactiveAccount(t *testing.T) {
	handler, fakeRepo, e := setupHandler()

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

	loginBody := `{"email":"inactive@example.com","password":"secret"}`
	reqLogin := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(loginBody))
	reqLogin.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	recLogin := httptest.NewRecorder()
	cLogin := e.NewContext(reqLogin, recLogin)

	err = handler.Login(cLogin)
	if err != nil {
		e.HTTPErrorHandler(err, cLogin)
	}

	assert.Equal(t, http.StatusForbidden, recLogin.Code)
	var resp map[string]interface{}
	assert.NoError(t, json.Unmarshal(recLogin.Body.Bytes(), &resp))
	assert.Equal(t, "Account is inactive", resp["error"])
}

func TestMeEndpoint(t *testing.T) {
	handler, fakeRepo, e := setupHandler()

	// seed the user directly into the fakeRepo slice
	// This bypasses the service but ensures we are testing the HANDLER's
	// ability to retrieve data based on claims.
	seededUser := user.User{
		ID:     1,
		Email:  "me@example.com",
		Role:   "user",
		Status: "active",
	}
	fakeRepo.Users = append(fakeRepo.Users, seededUser)

	// prepare echo request/recorder
	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// inject JWT claims manually (simulating middleware)
	claims := &auth.Claims{
		UserID: seededUser.ID,
		Email:  seededUser.Email,
		Role:   seededUser.Role,
		Status: seededUser.Status,
	}
	c.Set("user", claims)

	// call /me handler and assert
	if assert.NoError(t, handler.Me(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp dto.UserResponse
		assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		assert.Equal(t, seededUser.Email, resp.Email)
		assert.Equal(t, seededUser.ID, resp.ID)
	}
}

func TestUserRegistration_Duplicate(t *testing.T) {
	handler, fakeRepo, e := setupHandler()

	//  seed an existing user
	email := "duplicate@example.com"
	fakeRepo.Users = append(fakeRepo.Users, user.User{
		Email: email,
	})

	// try to register with the same email
	reqBody := `{"email":"` + email + `","password":"newpassword"}`
	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// call the handler and pass the error to the translator
	err := handler.Register(c)
	if err != nil {
		e.HTTPErrorHandler(err, c)
	}

	// assert: Expect 409 Conflict (assuming the translator maps it there)
	assert.Equal(t, http.StatusConflict, rec.Code)

	var resp map[string]interface{}
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "user already exists", resp["error"])
}

func TestUserRegistration_ContextCancelled(t *testing.T) {
	handler, _, e := setupHandler()

	reqBody := `{"email":"cancel@example.com","password":"password"}`
	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	// Create a context that is already cancelled
	ctx, cancel := context.WithCancel(req.Context())
	cancel() // Cancel it immediately
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute
	err := handler.Register(c)

	// Assert: The error should be context.Canceled or wrapped by it
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")
}

func TestAdminScrubUser_Handler(t *testing.T) {
	handler, fakeRepo, e := setupHandler()

	// Seed a user
	targetID := int64(5)
	fakeRepo.Users = append(fakeRepo.Users, user.User{
		ID:     targetID,
		Email:  "staff@company.com",
		Status: "active",
	})

	t.Run("successful scrub via DELETE request", func(t *testing.T) {
		// Prepare request for DELETE /admin/users/5
		req := httptest.NewRequest(http.MethodDelete, "/admin/users/5", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("5")

		// Call the handler
		if assert.NoError(t, handler.ScrubUser(c)) {
			assert.Equal(t, http.StatusNoContent, rec.Code)

			// Verify fake repo state
			u := fakeRepo.Users[0]
			assert.Equal(t, "deleted", u.Status)
			assert.Contains(t, u.Email, "deleted_5")
		}
	})

	t.Run("bad request on invalid ID string", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/admin/users/abc", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("abc")

		err := handler.ScrubUser(c)
		assert.Error(t, err)
		// Echo.NewHTTPError usually returns a 400 here
	})
}
