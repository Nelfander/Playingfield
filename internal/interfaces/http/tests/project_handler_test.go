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

func TestUpdateProject(t *testing.T) {
	handler, fakeRepo := setupProjectHandler()
	e := echo.New()

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
	c.SetPath("/projects/:id")
	c.SetParamNames("id")
	c.SetParamValues(fmt.Sprintf("%d", p.ID))

	// owner as the requester
	c.Set("user", &auth.Claims{UserID: ownerID})

	// execute the handler
	if assert.NoError(t, handler.Update(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)

		// verify the response
		var resp map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &resp)
		assert.Equal(t, "New Shiny Name", resp["name"])

		// verify the Fake DB state
		updated, _ := fakeRepo.GetByID(context.Background(), p.ID)
		assert.Equal(t, "New Shiny Name", updated.Name)
		assert.Equal(t, "Updated through the API", updated.Description)
	}
}

func TestListProjects(t *testing.T) {
	handler, fakeRepo := setupProjectHandler()
	e := echo.New()

	//  Seed the fake database with some projects
	ownerID := int64(100)
	fakeRepo.CreateProject(context.Background(), projects.Project{Name: "Project 1", OwnerID: ownerID})
	fakeRepo.CreateProject(context.Background(), projects.Project{Name: "Project 2", OwnerID: ownerID})
	fakeRepo.CreateProject(context.Background(), projects.Project{Name: "Other User Project", OwnerID: 999})

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
	p, _ := fakeRepo.CreateProject(context.Background(), projects.Project{
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

func TestAddUserToProject(t *testing.T) {
	handler, fakeRepo := setupProjectHandler()
	e := echo.New()

	ownerID := int64(100)
	targetUserID := int64(200)

	// create the project in the fake repo
	p, err := fakeRepo.CreateProject(context.Background(), projects.Project{
		Name:    "Collab Project",
		OwnerID: ownerID,
	})
	assert.NoError(t, err)

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

	// execute and assert status
	err = handler.AddUserToProject(c)
	assert.NoError(t, err)

	// stop if status is not 200 to see why it failed
	if rec.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK but got %d. Body: %s", rec.Code, rec.Body.String())
	}

	// verify json Response
	var resp map[string]string
	json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "User added successfully", resp["message"])

	// verify repo state
	members, err := fakeRepo.ListUsersInProject(context.Background(), p.ID)
	assert.NoError(t, err)

	// guard against panic only check index 0 if len is 1
	if assert.Equal(t, 1, len(members), "There should be exactly one member added") {
		assert.Equal(t, targetUserID, members[0].ID)

		// role check
		assert.Equal(t, "member", members[0].Role)
	}
}

func TestAddUserToProjectUnauthorized(t *testing.T) {
	handler, fakeRepo := setupProjectHandler()
	e := echo.New()

	// 3 users , 1 owner , 1 the hacker and 1 is the target
	ownerID := int64(100)
	targetUserID := int64(200)
	hackerID := int64(666)

	// create a project owned by user 100(ownerid)
	p, _ := fakeRepo.CreateProject(context.Background(), projects.Project{
		Name:    "is 666 considered evil?",
		OwnerID: ownerID,
	})

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

	// check for 403 Forbidden
	if err != nil {
		he, ok := err.(*echo.HTTPError)
		assert.True(t, ok)
		assert.Equal(t, http.StatusForbidden, he.Code)
	} else {
		assert.Equal(t, http.StatusForbidden, rec.Code)
	}

	// verify that the repository remains empty
	members, _ := fakeRepo.ListUsersInProject(context.Background(), p.ID)
	assert.Equal(t, 0, len(members), "the hacker should not have been able to add any members")

}

func TestRemoveUserFromProject(t *testing.T) {
	handler, fakeRepo := setupProjectHandler()
	e := echo.New()

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

	// execute the handler
	if assert.NoError(t, handler.RemoveUserFromProject(c)) {
		//  response is successful
		assert.Equal(t, http.StatusOK, rec.Code)

		// check that the fake DB is now empty for this project
		finalMembers, _ := fakeRepo.ListUsersInProject(context.Background(), p.ID)
		assert.Equal(t, 0, len(finalMembers), "The member list should be empty after removal")
	}
}

func TestRemoveUserFromProject_Unauthorized(t *testing.T) {
	handler, fakeRepo := setupProjectHandler()
	e := echo.New()

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
		he, ok := err.(*echo.HTTPError)
		assert.True(t, ok)
		assert.Equal(t, http.StatusForbidden, he.Code)
	} else {
		assert.Equal(t, http.StatusForbidden, rec.Code)
	}

	// ensure the user was NOT actually removed from the repo
	members, _ := fakeRepo.ListUsersInProject(context.Background(), p.ID)
	assert.Equal(t, 1, len(members), "The user should still be in the project!")
}

func TestAddUserToProject_Duplicate(t *testing.T) {
	handler, fakeRepo := setupProjectHandler()
	e := echo.New()

	ownerID := int64(100)
	targetUserID := int64(200)

	// create project
	p, _ := fakeRepo.CreateProject(context.Background(), projects.Project{
		Name:    "Duplicate Test Project",
		OwnerID: ownerID,
	})

	// manually add the user once via the repo
	_ = fakeRepo.AddUserToProject(context.Background(), p.ID, targetUserID, "member")

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

	// if handler returns the error directly to Echo
	if err != nil {
		assert.Contains(t, err.Error(), "already a member")
	} else {
		// if handler catches the error and writes to recorder
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "already a member")
	}
}
