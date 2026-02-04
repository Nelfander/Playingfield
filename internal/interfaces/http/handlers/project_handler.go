package handlers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/nelfander/Playingfield/internal/domain/projects"
	"github.com/nelfander/Playingfield/internal/infrastructure/auth"
)

type ProjectHandler struct {
	service *projects.Service
}

type AddUserToProjectRequest struct {
	ProjectID int64  `json:"project_id"`
	UserID    int64  `json:"user_id"`
	Role      string `json:"role"`
}

/*
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
*/
func NewProjectHandler(service *projects.Service) *ProjectHandler {
	return &ProjectHandler{service: service}
}

// POST /projects
func (h *ProjectHandler) Create(c echo.Context) error {
	var req struct {
		Name           string `json:"name"`
		Description    string `json:"description"`
		AssignedUserID string `json:"assigned_user_id"` // matches React frontend
	}

	if err := c.Bind(&req); err != nil {
		return err
	}

	claims, ok := c.Get("user").(*auth.Claims)
	if !ok || claims == nil {
		return echo.ErrUnauthorized
	}

	project, err := h.service.CreateProject(c.Request().Context(), req.Name, req.Description, claims.UserID)
	if err != nil {
		return err // errors.go handles conflicts/duplicates
	}

	if req.AssignedUserID != "" {
		targetUserID, parseErr := strconv.ParseInt(req.AssignedUserID, 10, 64)
		if parseErr == nil {
			_ = h.service.AddUserToProject(c.Request().Context(), 0, project.ID, targetUserID, "member")
		}
	}

	return c.JSON(http.StatusCreated, project)
}

// PUT /projects/:id
func (h *ProjectHandler) Update(c echo.Context) error {
	// parse project id from the url (/projects/:id)
	projectID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid project id")
	}

	// bind JSON body
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := c.Bind(&req); err != nil {
		return err
	}

	// get the requester's id from the context (auth middleware)
	userClaims := c.Get("user").(*auth.Claims)
	requesterID := userClaims.UserID

	if userClaims == nil {
		return echo.ErrUnauthorized
	}

	// call the Service
	updatedProject, err := h.service.UpdateProject(c.Request().Context(), requesterID, projectID, req.Name, req.Description)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, updatedProject)
}

// GET /projects
func (h *ProjectHandler) List(c echo.Context) error {
	claims, ok := c.Get("user").(*auth.Claims)
	if !ok || claims == nil {
		return echo.ErrUnauthorized
	}

	list, err := h.service.ListProjects(c.Request().Context(), claims.UserID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, list)
}

// DELETE /projects/:id
func (h *ProjectHandler) DeleteProject(c echo.Context) error {
	projectID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid project id")
	}

	claims, ok := c.Get("user").(*auth.Claims)
	if !ok || claims == nil {
		return echo.ErrUnauthorized
	}

	if err := h.service.DeleteProject(c.Request().Context(), projectID, claims.UserID); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *ProjectHandler) AddUserToProject(c echo.Context) error {
	var req AddUserToProjectRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	// extract requester's id from the jwt claims set by middleware
	claims, ok := c.Get("user").(*auth.Claims)
	if !ok || claims == nil {
		return echo.ErrUnauthorized
	}

	err := h.service.AddUserToProject(c.Request().Context(), claims.UserID, req.ProjectID, req.UserID, req.Role)
	if err != nil {
		return err // all cases already in errors.go (duplicate,authorized or 500)
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "User added successfully"})
}

func (h *ProjectHandler) ListUsersInProject(c echo.Context) error {
	projectIDParam := c.QueryParam("project_id")
	if projectIDParam == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "project_id is required")
	}

	projectID, err := strconv.ParseInt(projectIDParam, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid project_id format")
	}

	// call the service
	users, err := h.service.ListUsersInProject(c.Request().Context(), projectID)
	if err != nil {
		return err
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
			Role:  u.Role,
		})
	}
	return c.JSON(http.StatusOK, resp)
}

// DELETE /projects/members
func (h *ProjectHandler) RemoveUserFromProject(c echo.Context) error {

	var req struct {
		ProjectID int64 `json:"project_id"`
		UserID    int64 `json:"user_id"`
	}

	if err := c.Bind(&req); err != nil {
		return err
	}

	claims, ok := c.Get("user").(*auth.Claims)
	if !ok || claims == nil {
		return echo.ErrUnauthorized
	}

	if err := h.service.RemoveUserFromProject(c.Request().Context(), claims.UserID, req.ProjectID, req.UserID); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, echo.Map{"status": "user removed"})
}

// GET /projects/:id
func (h *ProjectHandler) GetByID(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid project id")
	}

	claims, ok := c.Get("user").(*auth.Claims)
	if !ok || claims == nil {
		return echo.ErrUnauthorized
	}

	project, err := h.service.GetProject(c.Request().Context(), id)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, project)
}
