package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/nelfander/Playingfield/internal/domain/projects"
	"github.com/nelfander/Playingfield/internal/infrastructure/auth"
	"github.com/nelfander/Playingfield/internal/interfaces/http/handlers"
	"github.com/stretchr/testify/assert"
)

// setupProjectHandler creates the environment for project tests
func setupProjectHandler() (*handlers.ProjectHandler, *projects.FakeRepository) {
	fakeRepo := projects.NewFakeRepository()
	//  nil for the store/queries and nil for the hub
	// because the Service now calls s.repo and has a nil-check for the hub.
	service := projects.NewService(fakeRepo, nil)
	handler := handlers.NewProjectHandler(service)
	return handler, fakeRepo
}

func TestCreateProject(t *testing.T) {
	handler, _ := setupProjectHandler()
	e := echo.New()

	reqBody := `{"name":"New Portfolio","description":"My awesome work"}`
	req := httptest.NewRequest(http.MethodPost, "/projects", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	claims := &auth.Claims{
		UserID: 100,
		Email:  "owner@example.com",
	}
	c.Set("user", claims)

	if assert.NoError(t, handler.Create(c)) {
		assert.Equal(t, http.StatusCreated, rec.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		assert.NoError(t, err)

		assert.Equal(t, "New Portfolio", resp["name"])
		assert.Equal(t, float64(100), resp["owner_id"])
	}
}

func TestListProjects(t *testing.T) {
	handler, fakeRepo := setupProjectHandler()
	e := echo.New()

	//  Seed the fake database with some projects
	ownerID := int64(100)
	fakeRepo.Create(context.Background(), projects.Project{Name: "Project 1", OwnerID: ownerID})
	fakeRepo.Create(context.Background(), projects.Project{Name: "Project 2", OwnerID: ownerID})
	fakeRepo.Create(context.Background(), projects.Project{Name: "Other User Project", OwnerID: 999})

	//  Setup Request
	req := httptest.NewRequest(http.MethodGet, "/projects", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	//  Mock Authentication for User 100
	claims := &auth.Claims{UserID: ownerID}
	c.Set("user", claims)

	//  Execute and Assert
	if assert.NoError(t, handler.List(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp []map[string]interface{}
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		assert.NoError(t, err)

		assert.Equal(t, 2, len(resp))
		assert.Equal(t, "Project 1", resp[0]["name"])
		assert.Equal(t, "Project 2", resp[1]["name"])
	}
}

func TestDeleteProject_Security(t *testing.T) {
	handler, fakeRepo := setupProjectHandler()
	e := echo.New()

	// create a project owned by User 100
	ownerID := int64(100)
	p, _ := fakeRepo.Create(context.Background(), projects.Project{
		Name:    "Owner's Secret Project",
		OwnerID: ownerID,
	})

	// another user (user 200) tries to delete user 100's project
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/projects/:id")
	c.SetParamNames("id")
	c.SetParamValues(fmt.Sprintf("%d", p.ID))

	// mock authentication for the WRONG user (ID 200)
	hackerClaims := &auth.Claims{UserID: 200}
	c.Set("user", hackerClaims)

	err := handler.DeleteProject(c)

	if err != nil {
		he, ok := err.(*echo.HTTPError)
		assert.True(t, ok)
		assert.Equal(t, http.StatusForbidden, he.Code)
	} else {
		assert.Equal(t, http.StatusForbidden, rec.Code)
	}
	savedProject, _ := fakeRepo.GetByID(context.Background(), p.ID)
	assert.NotNil(t, savedProject, "The project should not have been deleted!")
}
