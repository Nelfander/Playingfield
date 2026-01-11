package user

import "time"

type User struct {
	ID           int64
	Email        string
	PasswordHash string
	Role         string
	Status       string
	CreatedAt    time.Time
}
