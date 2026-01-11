package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/nelfander/Playingfield/internal/domain/projects"
	"github.com/nelfander/Playingfield/internal/infrastructure/auth"
)

type ProjectHandler struct {
	service *projects.Service
}

func NewProjectHandler(service *projects.Service) *ProjectHandler {
	return &ProjectHandler{service: service}
}

// POST /projects
func (h *ProjectHandler) Create(c echo.Context) error {
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request"})
	}

	// 🔑 Identity comes from JWT
	claims, ok := c.Get("user").(*auth.Claims)
	if !ok || claims == nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "unauthorized"})
	}
	ownerID := claims.UserID

	project, err := h.service.CreateProject(c.Request().Context(), req.Name, req.Description, ownerID)
	if err != nil {
		fmt.Println("DEBUG: create project error:", err) // Keep this for logging

		// Properly handle the duplicate-name error
		if strings.Contains(err.Error(), "already have a project with the name") {
			return c.JSON(http.StatusConflict, echo.Map{
				"error": err.Error(),
			})
		}

		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to create project"})
	}

	return c.JSON(http.StatusCreated, project)
}

// GET /projects?owner_id=123
func (h *ProjectHandler) List(c echo.Context) error {
	claims, ok := c.Get("user").(*auth.Claims)
	if !ok || claims == nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "unauthorized"})
	}

	ownerID := claims.UserID

	projects, err := h.service.ListProjects(c.Request().Context(), ownerID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to fetch projects"})
	}

	return c.JSON(http.StatusOK, projects)
}
