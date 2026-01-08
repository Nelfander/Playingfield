package postgres

import (
	"context"

	"github.com/nelfander/Playingfield/internal/domain/user"
	"github.com/nelfander/Playingfield/internal/infrastructure/auth"
)

func SeedAdminUser(ctx context.Context, userRepo user.Repository) error {
	adminEmail := "admin@example.com"
	adminPassword := "supersecret"

	_, err := userRepo.GetByEmail(ctx, adminEmail)
	if err == nil {
		// Admin already exists
		return nil
	}

	hash, err := auth.HashPassword(adminPassword)
	if err != nil {
		return err
	}

	admin := user.User{
		Email:        adminEmail,
		PasswordHash: hash,
		Role:         "admin",
	}

	_, err = userRepo.Create(ctx, admin)
	return err
}
