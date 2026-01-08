package dto

import (
	"time"

	"github.com/nelfander/Playingfield/internal/domain/user"
)

type UserResponse struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

// MapUser maps domain user -> UserResponse DTO
func MapUser(u *user.User) UserResponse {
	return UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
	}
}
