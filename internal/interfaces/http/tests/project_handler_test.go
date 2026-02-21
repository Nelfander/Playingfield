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
	custom "github.com/nelfander/Playingfield/internal/interfaces/http"
	"github.com/nelfander/Playingfield/internal/interfaces/http/handlers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupProjectHandler creates the environment for project tests
func setupProjectHandler() (*handlers.ProjectHandler, *projects.FakeRepository, *echo.Echo) {
	e := echo.New()
	e.HTTPErrorHandler = custom.CustomHTTPErrorHandler

	fakeRepo := projects.NewFakeRepository()
	service := projects.NewService(fakeRepo, nil)
	handler := handlers.NewProjectHandler(service)
	return handler, fakeRepo, e
}

func TestCreateProject(t *testing.T) {
	handler, _, e := setupProjectHandler()

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

	// pass errors to the handler to ensure Translator runs
	err := handler.Create(c)
	if err != nil {
		e.HTTPErrorHandler(err, c)
	}

	assert.Equal(t, http.StatusCreated, rec.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "New Portfolio", resp["name"])
	assert.Equal(t, float64(100), resp["owner_id"])
}

func TestUpdateProject(t *testing.T) {
	handler, fakeRepo, e := setupProjectHandler()

	ownerID := int64(100)
	// create a project to update
	p, _ := fakeRepo.CreateProject(context.Background(), projects.Project{
		Name:        "Old Project Name",
		Description: "Old Description",
		OwnerID:     ownerID,
	})

	// prepare the update payload
	input := map[string]interface{}{
		"name":        "New Shiny Name",
		"description": "Updated through the API",
	}
	body, _ := json.Marshal(input)

	req := httptest.NewRequest(http.MethodPut, "/projects/"+fmt.Sprintf("%d", p.ID), strings.NewReader(string(body)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// url param to match the route /projects/:id
	// mimics how Echo's router identifies which project to update.
	c.SetPath("/projects/:id")
	c.SetParamNames("id")
	c.SetParamValues(fmt.Sprintf("%d", p.ID))

	// owner as the requester
	c.Set("user", &auth.Claims{UserID: ownerID})

	// pass errors to the handler to ensure Translator runs
	err := handler.Update(c)
	if err != nil {
		e.HTTPErrorHandler(err, c)
	}

	assert.Equal(t, http.StatusOK, rec.Code)

	// verify the response contains the updated description
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "New Shiny Name", resp["name"])
	assert.Equal(t, "Updated through the API", resp["description"])
}

func TestListProjects(t *testing.T) {
	handler, fakeRepo, e := setupProjectHandler()
	//  Seed the fake database with some projects
	ownerID := int64(100)
	if _, err := fakeRepo.CreateProject(context.Background(), projects.Project{Name: "Project 1", OwnerID: ownerID}); err != nil {
		t.Fatalf("failed to create test project 1: %v", err)
	}
	if _, err := fakeRepo.CreateProject(context.Background(), projects.Project{Name: "Project 2", OwnerID: ownerID}); err != nil {
		t.Fatalf("failed to create test project 2: %v", err)
	}
	if _, err := fakeRepo.CreateProject(context.Background(), projects.Project{Name: "Other User Project", OwnerID: 999}); err != nil {
		t.Fatalf("failed to create other user's test project: %v", err)
	}

	//  Setup Request
	req := httptest.NewRequest(http.MethodGet, "/projects", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	//  Mock Authentication for User 100
	claims := &auth.Claims{UserID: ownerID}
	c.Set("user", claims)

	// capture the error and pass it to the handler, even on the happy path.
	err := handler.List(c)
	if err != nil {
		e.HTTPErrorHandler(err, c)
	}

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp []map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)

	assert.Equal(t, 2, len(resp))
	assert.Equal(t, "Project 1", resp[0]["name"])
	assert.Equal(t, "Project 2", resp[1]["name"])
}

func TestDeleteProject_Security(t *testing.T) {
	handler, fakeRepo, e := setupProjectHandler()

	// create a project owned by User 100
	ownerID := int64(100)
	p, _ := fakeRepo.CreateProject(context.Background(), projects.Project{
		Name:    "Owner's Secret Project",
		OwnerID: ownerID,
	})

	// another user (user 200) tries to delete user 100's project
	req := httptest.NewRequest(http.MethodDelete, "/projects/"+fmt.Sprintf("%d", p.ID), nil)
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
		e.HTTPErrorHandler(err, c) // this turns the error into the 403 response
	}

	assert.Equal(t, http.StatusForbidden, rec.Code, "Expected 403 for unauthorized delete")
}

func TestAddUserToProject(t *testing.T) {
	handler, fakeRepo, e := setupProjectHandler()

	ownerID := int64(100)
	targetUserID := int64(200)

	// create the project in the fake repo
	p, _ := fakeRepo.CreateProject(context.Background(), projects.Project{
		Name:    "Collab Project",
		OwnerID: ownerID,
	})

	input := map[string]interface{}{
		"project_id": p.ID,
		"user_id":    targetUserID,
		"role":       "member",
	}
	body, _ := json.Marshal(input)

	req := httptest.NewRequest(http.MethodPost, "/projects/members", strings.NewReader(string(body)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	//ensure the owner is the requester
	c.Set("user", &auth.Claims{UserID: ownerID})

	err := handler.AddUserToProject(c)
	if err != nil {
		e.HTTPErrorHandler(err, c)
	}

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAddUserToProjectUnauthorized(t *testing.T) {
	handler, fakeRepo, e := setupProjectHandler()
	svc := projects.NewService(fakeRepo, nil)

	// 3 users , 1 owner , 1 hacker and 1 target
	ownerID := int64(100)
	targetUserID := int64(200)
	hackerID := int64(666)

	// create a project owned by user 100(ownerid)
	p, _ := svc.CreateProject(context.Background(), "Secure Project", "Desc", ownerID)

	// prepare the json payload to add user 200
	input := map[string]interface{}{
		"project_id": p.ID,
		"user_id":    targetUserID,
		"role":       "member",
	}
	body, _ := json.Marshal(input)
	req := httptest.NewRequest(http.MethodPost, "/projects/members", strings.NewReader(string(body)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// logged in as the hacker(or 666)
	hackerClaims := &auth.Claims{UserID: hackerID}
	c.Set("user", hackerClaims)

	// this should be blocked
	err := handler.AddUserToProject(c)
	if err != nil {
		e.HTTPErrorHandler(err, c)
	}

	assert.Equal(t, http.StatusForbidden, rec.Code)

	// VERIFICATION: Members count should still be 1 (only the owner)
	members, _ := fakeRepo.ListUsersInProject(context.Background(), p.ID)
	assert.Equal(t, 1, len(members), "Only the owner should be in the project")
	assert.Equal(t, ownerID, members[0].ID)

}

func TestRemoveUserFromProject(t *testing.T) {
	handler, fakeRepo, e := setupProjectHandler()

	//  Creates a project and pre-adds a member
	ownerID := int64(100)
	targetUserID := int64(200)

	p, _ := fakeRepo.CreateProject(context.Background(), projects.Project{
		Name:    "Project to Clean Up",
		OwnerID: ownerID,
	})

	// manually inject the user into the fake repo so they exist to be removed
	_ = fakeRepo.AddUserToProject(context.Background(), p.ID, targetUserID, "member")

	// verify the user is actually there before it starts
	initialMembers, _ := fakeRepo.ListUsersInProject(context.Background(), p.ID)
	assert.Equal(t, 1, len(initialMembers))

	//  prepare the delete request
	input := map[string]interface{}{
		"project_id": p.ID,
		"user_id":    targetUserID,
	}
	body, _ := json.Marshal(input)

	req := httptest.NewRequest(http.MethodDelete, "/projects/members", strings.NewReader(string(body)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// mock authentication: the owner is the one performing the removal
	claims := &auth.Claims{UserID: ownerID}
	c.Set("user", claims)

	// execute with the Error Bridge
	err := handler.RemoveUserFromProject(c)
	if err != nil {
		e.HTTPErrorHandler(err, c)
	}

	// API Verification
	assert.Equal(t, http.StatusOK, rec.Code)

	// repo verification (if user is actually removed)
	finalMembers, _ := fakeRepo.ListUsersInProject(context.Background(), p.ID)
	assert.Equal(t, 0, len(finalMembers), "The member list should be empty in the DB after removal")
}

func TestRemoveUserFromProject_Unauthorized(t *testing.T) {
	handler, fakeRepo, e := setupProjectHandler()

	ownerID := int64(100)
	hackerID := int64(666) // the unauthorized user
	targetUserID := int64(200)

	p, _ := fakeRepo.CreateProject(context.Background(), projects.Project{
		Name:    "Secure Project",
		OwnerID: ownerID,
	})

	// add the user to the project
	_ = fakeRepo.AddUserToProject(context.Background(), p.ID, targetUserID, "member")

	input := map[string]interface{}{
		"project_id": p.ID,
		"user_id":    targetUserID,
	}
	body, _ := json.Marshal(input)

	req := httptest.NewRequest(http.MethodDelete, "/projects/members", strings.NewReader(string(body)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	//  logged in as the "hacker", not the owner
	claims := &auth.Claims{UserID: hackerID}
	c.Set("user", claims)

	// system should reject this
	err := handler.RemoveUserFromProject(c)
	if err != nil {
		e.HTTPErrorHandler(err, c)
	}

	assert.Equal(t, http.StatusForbidden, rec.Code, "Expected 403 for unauthorized member removal")

	// ensure the user was NOT actually removed from the repo
	members, _ := fakeRepo.ListUsersInProject(context.Background(), p.ID)
	assert.Equal(t, 1, len(members), "The user should still be in the project!")
}

func TestAddUserToProject_Duplicate(t *testing.T) {
	handler, fakeRepo, e := setupProjectHandler()
	svc := projects.NewService(fakeRepo, nil)

	ownerID := int64(100)
	targetUserID := int64(200)

	// create project
	p, _ := svc.CreateProject(context.Background(), "Duplicate Test", "Desc", ownerID)

	// add the user the first time via service logic
	_ = svc.AddUserToProject(context.Background(), ownerID, p.ID, targetUserID, "member")

	// try to add the same user again via the Handler
	input := map[string]interface{}{
		"project_id": p.ID,
		"user_id":    targetUserID,
		"role":       "member",
	}
	body, _ := json.Marshal(input)
	req := httptest.NewRequest(http.MethodPost, "/projects/members", strings.NewReader(string(body)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", &auth.Claims{UserID: ownerID})

	// assert that it fails
	err := handler.AddUserToProject(c)
	if err != nil {
		e.HTTPErrorHandler(err, c)
	}

	// assert specific 409 Conflict
	assert.Equal(t, http.StatusConflict, rec.Code)
	assert.Contains(t, rec.Body.String(), "already a member")
}
