package handlers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/nelfander/Playingfield/internal/domain/messages"
	"github.com/nelfander/Playingfield/internal/infrastructure/auth"
)

type ChatHandler struct {
	service messages.ChatService // Use the interface name, no asterisk!
}

func NewChatHandler(service messages.ChatService) *ChatHandler {
	return &ChatHandler{service: service}
}

// GET /projects/:id/messages
func (h *ChatHandler) GetProjectHistory(c echo.Context) error {
	projectID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid project id")
	}

	history, err := h.service.GetProjectHistory(c.Request().Context(), projectID)
	if err != nil {
		return err // translator will handle it
	}

	return c.JSON(http.StatusOK, history)
}

// GET /messages/direct/:other_id
func (h *ChatHandler) GetDMHistory(c echo.Context) error {
	claims, ok := c.Get("user").(*auth.Claims)
	if !ok || claims == nil {
		return echo.ErrUnauthorized
	}
	myID := claims.UserID

	otherUserID, err := strconv.ParseInt(c.Param("other_id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user id")
	}

	history, err := h.service.GetDMHistory(c.Request().Context(), myID, otherUserID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, history)
}
