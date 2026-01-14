package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/nelfander/Playingfield/internal/domain/projects"
	"github.com/nelfander/Playingfield/internal/infrastructure/auth"
	"github.com/nelfander/Playingfield/internal/infrastructure/postgres/sqlc"
)

type ProjectHandler struct {
	service *projects.Service
}

type AddUserToProjectRequest struct {
	ProjectID int64  `json:"project_id"`
	UserID    int64  `json:"user_id"`
	Role      string `json:"role"`
}

type ProjectUserResponse struct {
	ID        int64  `json:"id"`
	ProjectID int64  `json:"project_id"`
	UserID    int64  `json:"user_id"`
	Role      string `json:"role"`
}

func ProjectUserToResponse(p sqlc.ProjectUser) ProjectUserResponse {
	return ProjectUserResponse{
		ID:        p.ID,
		ProjectID: p.ProjectID,
		UserID:    p.UserID,
		Role:      p.Role.String,
	}
}

func NewProjectHandler(service *projects.Service) *ProjectHandler {
	return &ProjectHandler{service: service}
}

// POST /projects
func (h *ProjectHandler) Create(c echo.Context) error {
	var req struct {
		Name           string `json:"name"`
		Description    string `json:"description"`
		AssignedUserID string `json:"assigned_user_id"` // Matches React frontend
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request"})
	}

	claims, ok := c.Get("user").(*auth.Claims)
	if !ok || claims == nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "unauthorized"})
	}
	ownerID := claims.UserID

	project, err := h.service.CreateProject(c.Request().Context(), req.Name, req.Description, ownerID)
	if err != nil {
		if strings.Contains(err.Error(), "already have a project with the name") {
			return c.JSON(http.StatusConflict, echo.Map{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to create project"})
	}

	if req.AssignedUserID != "" {
		targetUserID, parseErr := strconv.ParseInt(req.AssignedUserID, 10, 64)
		if parseErr == nil {
			_, _ = h.service.AddUserToProject(project.ID, targetUserID, "member")
		}
	}

	return c.JSON(http.StatusCreated, project)
}

// GET /projects
func (h *ProjectHandler) List(c echo.Context) error {
	claims, ok := c.Get("user").(*auth.Claims)
	if !ok || claims == nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "unauthorized"})
	}

	currentUserID := claims.UserID

	projects, err := h.service.ListProjects(c.Request().Context(), currentUserID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to fetch projects"})
	}

	return c.JSON(http.StatusOK, projects)
}

func (h *ProjectHandler) DeleteProject(c echo.Context) error {
	projectID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid project ID"})
	}

	userID := c.Get("user_id").(int64)

	err = h.service.DeleteProject(c.Request().Context(), projectID, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to delete project"})
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *ProjectHandler) AddUserToProject(c echo.Context) error {
	var req AddUserToProjectRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	projectUser, err := h.service.AddUserToProject(req.ProjectID, req.UserID, req.Role)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, ProjectUserToResponse(projectUser))
}

func (h *ProjectHandler) ListUsersInProject(c echo.Context) error {
	// get project ID from query param (or JSON if you prefer)
	projectIDParam := c.QueryParam("project_id")
	if projectIDParam == "" {
		return c.JSON(400, map[string]string{"error": "project_id is required"})
	}

	projectID, err := strconv.ParseInt(projectIDParam, 10, 64)
	if err != nil {
		return c.JSON(400, map[string]string{"error": "invalid project_id"})
	}

	// call the service
	users, err := h.service.ListUsersInProject(projectID)
	if err != nil {
		return c.JSON(500, map[string]string{"error": err.Error()})
	}

	// convert to JSON-friendly response
	type UserResponse struct {
		ID    int64  `json:"id"`
		Email string `json:"email"`
		Role  string `json:"role"`
	}

	var resp []UserResponse
	for _, u := range users {
		resp = append(resp, UserResponse{
			ID:    u.ID,
			Email: u.Email,
			Role:  u.Role.String,
		})
	}

	return c.JSON(200, resp)
}

func (h *ProjectHandler) RemoveUserFromProject(c echo.Context) error {

	type RemoveUserRequest struct {
		ProjectID int64 `json:"project_id"`
		UserID    int64 `json:"user_id"`
	}

	var req RemoveUserRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(400, map[string]string{"error": "invalid request"})
	}

	requesterIDInterface := c.Get("user_id")
	if requesterIDInterface == nil {
		return c.JSON(401, map[string]string{"error": "unauthorized"})
	}
	requesterID := requesterIDInterface.(int64)

	err := h.service.RemoveUserFromProject(requesterID, req.ProjectID, req.UserID)
	if err != nil {
		return c.JSON(403, map[string]string{"error": err.Error()})
	}

	return c.JSON(200, map[string]string{"status": "user removed"})
}
