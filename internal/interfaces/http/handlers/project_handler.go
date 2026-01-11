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

func NewProjectHandler(service *projects.Service) *ProjectHandler {
	return &ProjectHandler{service: service}
}

// POST /projects
func (h *ProjectHandler) Create(c echo.Context) error {
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		OwnerID     int64  `json:"owner_id"` // for now, pass manually
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request"})
	}

	// 🔑 Identity comes from JWT
	claims, ok := c.Get("user").(*auth.Claims)
	if !ok || claims == nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "unauthorized"})
	}

	project, err := h.service.CreateProject(
		c.Request().Context(),
		req.Name,
		req.Description,
		claims.UserID,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to create project"})
	}

	return c.JSON(http.StatusCreated, project)
}

// GET /projects?owner_id=123
func (h *ProjectHandler) List(c echo.Context) error {
	ownerIDStr := c.QueryParam("owner_id")
	ownerID, err := strconv.ParseInt(ownerIDStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid owner_id"})
	}

	projects, err := h.service.ListProjects(c.Request().Context(), ownerID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, projects)
}
