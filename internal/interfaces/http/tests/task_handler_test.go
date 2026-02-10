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
	"github.com/nelfander/Playingfield/internal/domain/tasks"
	"github.com/nelfander/Playingfield/internal/infrastructure/auth"
	custom "github.com/nelfander/Playingfield/internal/interfaces/http"
	"github.com/nelfander/Playingfield/internal/interfaces/http/handlers"
	"github.com/stretchr/testify/assert"
)

func setupTaskHandler() (*handlers.TaskHandler, *tasks.FakeRepository, *projects.FakeRepository, *echo.Echo) {
	e := echo.New()
	e.HTTPErrorHandler = custom.CustomHTTPErrorHandler

	fakeTaskRepo := tasks.NewFakeRepository()
	fakeProjRepo := projects.NewFakeRepository()

	// task service needs BOTH repos
	service := tasks.NewService(fakeTaskRepo, fakeProjRepo, nil, nil)
	handler := handlers.NewTaskHandler(service)

	return handler, fakeTaskRepo, fakeProjRepo, e
}

func TestCreateTask_Handler(t *testing.T) {
	handler, _, fakeProjRepo, e := setupTaskHandler()

	ownerID := int64(100)
	p, _ := fakeProjRepo.CreateProject(context.Background(), projects.Project{Name: "Test Proj", OwnerID: ownerID})

	t.Run("successful task creation", func(t *testing.T) {
		reqBody := fmt.Sprintf(`{"project_id":%d,"title":"Task from API","status":"todo"}`, p.ID)
		req := httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		c.Set("user", &auth.Claims{UserID: ownerID})

		if err := handler.CreateTask(c); err != nil {
			e.HTTPErrorHandler(err, c)
		}

		assert.Equal(t, http.StatusCreated, rec.Code)

		var resp map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &resp)
		assert.Equal(t, "Task from API", resp["title"])
	})
}

func TestUpdateTask_HandlerSecurity(t *testing.T) {
	handler, fakeTaskRepo, fakeProjRepo, e := setupTaskHandler()

	ownerID := int64(100)
	hackerID := int64(666)

	p, _ := fakeProjRepo.CreateProject(context.Background(), projects.Project{Name: "Secure", OwnerID: ownerID})
	task, _ := fakeTaskRepo.CreateTask(context.Background(), &tasks.Task{ProjectID: p.ID, Title: "Don't Touch"})

	t.Run("unauthorized user gets 403", func(t *testing.T) {
		reqBody := `{"status":"hacked"}`
		req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/tasks/%d", task.ID), strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		c.SetPath("/tasks/:id")
		c.SetParamNames("id")
		c.SetParamValues(fmt.Sprintf("%d", task.ID))
		c.Set("user", &auth.Claims{UserID: hackerID})

		err := handler.UpdateTask(c)
		if err != nil {
			e.HTTPErrorHandler(err, c)
		}

		// this confirms the Domain -> Service -> Handler -> Translator chain works
		assert.Equal(t, http.StatusForbidden, rec.Code)
	})
}

func TestListTasks_Handler(t *testing.T) {
	handler, fakeTaskRepo, fakeProjRepo, e := setupTaskHandler()
	ownerID := int64(100)

	// Seed 1 Project and 2 Tasks
	p, _ := fakeProjRepo.CreateProject(context.Background(), projects.Project{Name: "List Project", OwnerID: ownerID})

	// manually add the owner to the project members in the fake repo
	// because we are bypassing the ProjectService.CreateProject logic
	_ = fakeProjRepo.AddUserToProject(context.Background(), p.ID, ownerID, "owner")

	fakeTaskRepo.CreateTask(context.Background(), &tasks.Task{ProjectID: p.ID, Title: "Task A"})
	fakeTaskRepo.CreateTask(context.Background(), &tasks.Task{ProjectID: p.ID, Title: "Task B"})

	t.Run("member can list tasks", func(t *testing.T) {
		// Simulating GET /projects/:id/tasks
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/projects/%d/tasks", p.ID), nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		c.SetPath("/projects/:id/tasks")
		c.SetParamNames("id")
		c.SetParamValues(fmt.Sprintf("%d", p.ID))
		c.Set("user", &auth.Claims{UserID: ownerID})

		if err := handler.ListTaskByProject(c); err != nil {
			e.HTTPErrorHandler(err, c)
		}

		assert.Equal(t, http.StatusOK, rec.Code)

		var resp []map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &resp)
		assert.Len(t, resp, 2)
	})
}

func TestGetTaskHistory_Handler(t *testing.T) {
	handler, fakeTaskRepo, fakeProjRepo, e := setupTaskHandler()
	ownerID := int64(100)

	// setup: Project -> Task -> some activity
	p, _ := fakeProjRepo.CreateProject(context.Background(), projects.Project{Name: "Just a project", OwnerID: ownerID})
	// add the member so the authorization check passes
	_ = fakeProjRepo.AddUserToProject(context.Background(), p.ID, ownerID, "owner")
	task, _ := fakeTaskRepo.CreateTask(context.Background(), &tasks.Task{ProjectID: p.ID, Title: "tracked task"})

	// manually record an activity in the fake
	_ = fakeTaskRepo.RecordTaskActivity(context.Background(), &tasks.TaskActivity{
		TaskID: task.ID, UserID: ownerID, Action: "CREATED", Details: "Initial",
	})

	t.Run("authorized user can see history", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/tasks/%d/history", task.ID), nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		c.SetPath("/tasks/:id/history")
		c.SetParamNames("id")
		c.SetParamValues(fmt.Sprintf("%d", task.ID))
		c.Set("user", &auth.Claims{UserID: ownerID})

		if err := handler.GetTaskHistory(c); err != nil {
			e.HTTPErrorHandler(err, c)
		}

		assert.Equal(t, http.StatusOK, rec.Code)

		var resp []map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &resp)
		assert.NotEmpty(t, resp)
		assert.Equal(t, "CREATED", resp[0]["action"])
	})
}
