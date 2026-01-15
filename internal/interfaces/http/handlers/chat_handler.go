package handlers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/nelfander/Playingfield/internal/domain/messages"
	"github.com/nelfander/Playingfield/internal/infrastructure/auth"
)

type ChatHandler struct {
	service *messages.Service
}

func NewChatHandler(service *messages.Service) *ChatHandler {
	return &ChatHandler{service: service}
}

// GET /projects/:id/messages
func (h *ChatHandler) GetProjectHistory(c echo.Context) error {
	projectID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid project id"})
	}

	history, err := h.service.GetProjectHistory(c.Request().Context(), projectID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, history)
}

// GET /messages/direct/:other_id
func (h *ChatHandler) GetDMHistory(c echo.Context) error {
	claims, ok := c.Get("user").(*auth.Claims)
	if !ok || claims == nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "unauthorized"})
	}
	myID := claims.UserID

	otherUserID, err := strconv.ParseInt(c.Param("other_id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid user id"})
	}

	history, err := h.service.GetDMHistory(c.Request().Context(), myID, otherUserID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, history)
}
