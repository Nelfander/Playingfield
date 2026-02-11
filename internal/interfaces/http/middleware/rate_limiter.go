package middleware

import (
	"errors"
	"fmt"
	"log/slog"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/nelfander/Playingfield/internal/infrastructure/auth"
	"golang.org/x/time/rate"
)

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var (
	visitors             = make(map[string]*visitor)
	mu                   sync.RWMutex
	ErrRateLimitExceeded = errors.New("rate limit exceeded")
)

func init() {
	go cleanupVisitors()
}

func cleanupVisitors() {
	for {
		time.Sleep(time.Minute)
		mu.Lock()
		for ip, v := range visitors {
			if time.Since(v.lastSeen) > 10*time.Minute {
				delete(visitors, ip)
			}
		}
		mu.Unlock()
	}
}

// this is only so that tests can clear the state,
// in prod we do want visitors map to persist
func ResetVisitors() {
	mu.Lock()
	defer mu.Unlock()
	visitors = make(map[string]*visitor)
}

func RateLimitMiddleware(jwtManager *auth.JWTManager) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			var identifier string
			var limit rate.Limit
			var burst int

			// identify user and tier
			authHeader := c.Request().Header.Get("Authorization")

			// check if there is a valid token to upgrade the rate limit
			if strings.HasPrefix(authHeader, "Bearer ") {
				tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
				claims, err := jwtManager.VerifyToken(tokenStr)

				if err == nil {
					// authenticated = upgrade to 20 limit
					identifier = fmt.Sprintf("user_%d", claims.UserID)
					limit = rate.Limit(20)
					burst = 40

					// optimization: cache these for the next middlewares/handlers
					c.Set("user_id", claims.UserID)
					c.Set("user", claims)
				}
			}

			if identifier == "" {
				// if not an authenticated user, lower the rate
				ip, _, err := net.SplitHostPort(c.Request().RemoteAddr)
				if err != nil {
					identifier = c.Request().RemoteAddr
				} else {
					identifier = ip
				}
				limit = rate.Limit(5)
				burst = 10
			}

			slog.Debug("Rate limit check", "id", identifier)

			// limiter logic (high performance locking)
			mu.RLock()
			v, exists := visitors[identifier]
			mu.RUnlock()

			if !exists {
				mu.Lock()
				if v, exists = visitors[identifier]; !exists {
					v = &visitor{
						limiter:  rate.NewLimiter(limit, burst),
						lastSeen: time.Now(),
					}
					visitors[identifier] = v
				}
				mu.Unlock()
			}

			// update state
			mu.RLock()
			v.lastSeen = time.Now()
			mu.RUnlock()

			if !v.limiter.Allow() {
				slog.Warn("Rate limit exceeded", "id", identifier, "path", c.Path())
				return ErrRateLimitExceeded
			}

			return next(c)
		}
	}
}
