package http

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/nelfander/Playingfield/internal/domain/projects"
	"github.com/nelfander/Playingfield/internal/domain/tasks"
	"github.com/nelfander/Playingfield/internal/domain/user"
	"github.com/nelfander/Playingfield/internal/interfaces/http/middleware"
)

// CustomHTTPErrorHandler handles errors globally across the Echo instance
func CustomHTTPErrorHandler(err error, c echo.Context) {
	// Default response
	code := http.StatusInternalServerError
	message := "Internal Server Error"

	// Map Domain errors to HTTP status codes
	// Switch through domain errors
	switch {
	// --- User domain ---
	case errors.Is(err, user.ErrUserAlreadyExists):
		code = http.StatusConflict
		message = err.Error()

	case errors.Is(err, user.ErrInvalidCredentials):
		code = http.StatusUnauthorized
		message = "Invalid email or password"
		// The 'err' itself is already logged by slog, we see "wrong password"
		// from the service log and "invalid credentials" from the translator

	case errors.Is(err, user.ErrInactiveAccount):
		code = http.StatusForbidden
		message = "Account is inactive"

	// --- Task domain ---
	case errors.Is(err, tasks.ErrTaskNotFound):
		code = http.StatusNotFound
		message = "Task not found"

	case errors.Is(err, tasks.ErrUnauthorized):
		code = http.StatusForbidden
		message = "You do not have permission for this action"

	// --- Project domain ---
	case errors.Is(err, projects.ErrProjectNotFound):
		code = http.StatusNotFound
		message = "Project not found"

	case errors.Is(err, projects.ErrUnauthorized):
		// This covers both "only the project owner" and "not the owner of project"
		code = http.StatusForbidden
		message = err.Error()

	case errors.Is(err, projects.ErrAlreadyMember),
		errors.Is(err, projects.ErrDuplicateProject):
		// This covers "already a member" and "already have a project with the name"
		code = http.StatusConflict
		message = err.Error()

	// --- Rate Limiter ---
	case errors.Is(err, middleware.ErrRateLimitExceeded):
		code = http.StatusTooManyRequests
		message = "Slow down! You're sending too many requests."

	// --- Echo and System errors! ---
	default:
		var he *echo.HTTPError
		if errors.As(err, &he) {
			code = he.Code
			message = fmt.Sprintf("%v", he.Message)

		}
	}

	// only log 500s as actual "Error" level.
	// 400s are "Warn" or "Info" because they are usually the user's fault, not a system failure.
	if code >= 500 {
		slog.Error("server error",
			"err", err,
			"method", c.Request().Method,
			"path", c.Path(),
		)
	} else {
		slog.Warn("client error",
			"code", code,
			"err", err,
			"path", c.Path(),
		)
	}

	// send the response and LOG if the send itself fails
	if !c.Response().Committed {
		var respErr error
		if c.Request().Method == http.MethodHead {
			respErr = c.NoContent(code)
		} else {
			respErr = c.JSON(code, map[string]string{"error": message})
		}

		if respErr != nil {
			slog.Error("failed to send error response", "err", respErr)
		}
	}
}
