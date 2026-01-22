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
			_ = h.service.AddUserToProject(c.Request().Context(), 0, project.ID, targetUserID, "member")
		}
	}

	return c.JSON(http.StatusCreated, project)
}

func (h *ProjectHandler) Update(c echo.Context) error {
	// parse project id from the url (/projects/:id)
	idParam := c.Param("id")
	projectID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid project id")
	}

	// bind JSON body
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	// get the requester's id from the context (auth middleware)
	userClaims := c.Get("user").(*auth.Claims)
	requesterID := userClaims.UserID

	// call the Service
	updatedProject, err := h.service.UpdateProject(c.Request().Context(), requesterID, projectID, req.Name, req.Description)
	if err != nil {
		if strings.Contains(err.Error(), "unauthorized") {
			return echo.NewHTTPError(http.StatusForbidden, err.Error())
		}
		if strings.Contains(err.Error(), "not found") {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, updatedProject)
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

	claims, ok := c.Get("user").(*auth.Claims)
	if !ok || claims == nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "unauthorized"})
	}
	userID := claims.UserID

	err = h.service.DeleteProject(c.Request().Context(), projectID, userID)
	if err != nil {
		return c.JSON(http.StatusForbidden, echo.Map{"error": "You do not have permission to delete this project"})
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *ProjectHandler) AddUserToProject(c echo.Context) error {
	var req AddUserToProjectRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	// extract requester's id from the jwt claims set by middleware
	claims, ok := c.Get("user").(*auth.Claims)
	if !ok || claims == nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "unauthorized"})
	}
	requesterID := claims.UserID

	err := h.service.AddUserToProject(c.Request().Context(), requesterID, req.ProjectID, req.UserID, req.Role)
	if err != nil {
		// check if the error is due to a duplicate
		if strings.Contains(err.Error(), "already a member") {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		//  if authorization error return 403 Forbidden
		if strings.Contains(err.Error(), "unauthorized") {
			return c.JSON(http.StatusForbidden, map[string]string{"error": err.Error()})
		}
		// otherwise return 500
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "User added successfully"})
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
	users, err := h.service.ListUsersInProject(c.Request().Context(), projectID)
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
		roleStr := ""
		if u.Role != nil {
			if str, ok := u.Role.(string); ok {
				roleStr = str
			}
		}
		resp = append(resp, UserResponse{
			ID:    u.ID,
			Email: u.Email,
			Role:  roleStr,
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

	claims, ok := c.Get("user").(*auth.Claims)
	if !ok || claims == nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "unauthorized"})
	}
	requesterID := claims.UserID

	err := h.service.RemoveUserFromProject(requesterID, req.ProjectID, req.UserID)
	if err != nil {
		return c.JSON(403, map[string]string{"error": err.Error()})
	}

	return c.JSON(200, map[string]string{"status": "user removed"})
}

// GET /projects/:id
func (h *ProjectHandler) GetByID(c echo.Context) error {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid project id"})
	}

	claims, ok := c.Get("user").(*auth.Claims)
	if !ok || claims == nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "unauthorized"})
	}

	project, err := h.service.GetProject(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"error": "project not found"})
	}

	return c.JSON(http.StatusOK, project)
}
