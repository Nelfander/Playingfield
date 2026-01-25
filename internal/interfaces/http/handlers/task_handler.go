package handlers

import (
	"net/http"
	"strconv"
	"strings"

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
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request body"})
	}

	claims, ok := c.Get("user").(*auth.Claims)
	if !ok || claims == nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "unauthorized"})
	}

	task := tasks.Task{
		ProjectID:   req.ProjectID,
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
		AssignedTo:  req.AssignedTo,
	}

	created, err := h.service.CreateTask(c.Request().Context(), claims.UserID, task)
	if err != nil {
		if strings.Contains(err.Error(), "unauthorized") {
			return c.JSON(http.StatusForbidden, echo.Map{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, created)
}

// PUT /tasks/:id
func (h *TaskHandler) UpdateTask(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid task id"})
	}

	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Status      string `json:"status"`
		AssignedTo  *int64 `json:"assigned_to"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request body"})
	}

	claims := c.Get("user").(*auth.Claims)

	task := tasks.Task{
		ID:          id,
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
		AssignedTo:  req.AssignedTo,
	}

	updated, err := h.service.UpdateTask(c.Request().Context(), claims.UserID, task)
	if err != nil {
		if strings.Contains(err.Error(), "unauthorized") {
			return c.JSON(http.StatusForbidden, echo.Map{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, updated)
}

// GET /projects/:id/tasks
func (h *TaskHandler) ListTaskByProject(c echo.Context) error {
	projectID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid project id"})
	}

	claims := c.Get("user").(*auth.Claims)

	list, err := h.service.ListTasks(c.Request().Context(), claims.UserID, projectID)
	if err != nil {
		if strings.Contains(err.Error(), "unauthorized") {
			return c.JSON(http.StatusForbidden, echo.Map{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to fetch tasks"})
	}

	return c.JSON(http.StatusOK, list)
}

// DELETE /tasks/:id
func (h *TaskHandler) DeleteTask(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid task id"})
	}

	claims := c.Get("user").(*auth.Claims)

	err = h.service.DeleteTask(c.Request().Context(), claims.UserID, id)
	if err != nil {
		return c.JSON(http.StatusForbidden, echo.Map{"error": err.Error()})
	}

	return c.NoContent(http.StatusNoContent)
}

// GET /tasks/:id/history
func (h *TaskHandler) GetTaskHistory(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid task id"})
	}

	claims := c.Get("user").(*auth.Claims)

	history, err := h.service.GetTaskHistory(c.Request().Context(), claims.UserID, id)
	if err != nil {
		return c.JSON(http.StatusForbidden, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, history)
}
