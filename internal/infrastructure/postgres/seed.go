package postgres

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/nelfander/Playingfield/internal/domain/user"
	"github.com/nelfander/Playingfield/internal/infrastructure/auth"
)

// func for creating admin user, dont know if i keep it in prod
func SeedAdminUser(ctx context.Context, userRepo user.Repository) error {
	adminEmail := "admin@example.com"
	adminPassword := "supersecret"

	// check if admin exists
	_, err := userRepo.GetByEmail(ctx, adminEmail)
	if err == nil {
		slog.Debug("admin user already exists, skipping seed", "email", adminEmail)
		return nil
	}

	slog.Info("seeding admin user...", "email", adminEmail)

	hash, err := auth.HashPassword(adminPassword)
	if err != nil {
		return fmt.Errorf("seed: failed to hash admin password: %w", err)
	}

	admin := user.User{
		Email:        adminEmail,
		PasswordHash: hash,
		Role:         "admin",
		Status:       "active",
	}

	_, err = userRepo.Create(ctx, admin)
	if err != nil {
		return fmt.Errorf("seed: failed to create admin user: %w", err)
	}

	slog.Info("admin user seeded successfully", "email", adminEmail)
	return nil
}
