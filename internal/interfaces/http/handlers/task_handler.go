package handlers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/nelfander/Playingfield/internal/domain/tasks"
	"github.com/nelfander/Playingfield/internal/infrastructure/auth"
)

type TaskHandler struct {
	service *tasks.Service
}

func NewTaskHandler(service *tasks.Service) *TaskHandler {
	return &TaskHandler{service: service}
}

// POST /tasks
func (h *TaskHandler) CreateTask(c echo.Context) error {
	var req struct {
		ProjectID   int64  `json:"project_id"`
		Title       string `json:"title"`
		Description string `json:"description"`
		Status      string `json:"status"`
		AssignedTo  *int64 `json:"assigned_to"` // Pointer to allow null
	}

	if err := c.Bind(&req); err != nil {
		return err
	}

	claims, ok := c.Get("user").(*auth.Claims)
	if !ok || claims == nil {
		return echo.ErrUnauthorized
	}

	task := tasks.Task{
		ProjectID:   req.ProjectID,
		Title:       req.Title,
		Description: req.Description,
		Status:      "TODO",
		AssignedTo:  req.AssignedTo,
	}

	created, err := h.service.CreateTask(c.Request().Context(), claims.UserID, task)
	if err != nil {
		return err // automatically maps tasks.ErrUnauthorized to 403 via Translator
	}

	return c.JSON(http.StatusCreated, created)
}

// PUT /tasks/:id
func (h *TaskHandler) UpdateTask(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		// the error returned by strconv.ParseInt is a generic Go error
		// wrapping it in echo.NewHTTPError(400, "invalid id" to give context
		// and status code before translator handles it
		return echo.NewHTTPError(http.StatusBadRequest, "invalid task id")
	}

	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Status      string `json:"status"`
		AssignedTo  *int64 `json:"assigned_to"`
		Message     string `json:"message"`
	}

	if err := c.Bind(&req); err != nil {
		return err
	}

	claims := c.Get("user").(*auth.Claims)

	task := tasks.Task{
		ID:          id,
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
		AssignedTo:  req.AssignedTo,
	}

	updated, err := h.service.UpdateTask(c.Request().Context(), claims.UserID, task, req.Message)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, updated)
}

// GET /projects/:id/tasks
func (h *TaskHandler) ListTaskByProject(c echo.Context) error {
	projectID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid project id")
	}

	claims := c.Get("user").(*auth.Claims)

	list, err := h.service.ListTasks(c.Request().Context(), claims.UserID, projectID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, list)
}

// DELETE /tasks/:id
func (h *TaskHandler) DeleteTask(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid task id")
	}

	claims := c.Get("user").(*auth.Claims)

	if err := h.service.DeleteTask(c.Request().Context(), claims.UserID, id); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}

// GET /tasks/:id/history
func (h *TaskHandler) GetTaskHistory(c echo.Context) error {
	taskID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid task id")
	}

	claims := c.Get("user").(*auth.Claims)

	history, err := h.service.GetTaskHistory(c.Request().Context(), claims.UserID, taskID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, history)
}
