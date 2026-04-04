package app

import (
	"context"
	"errors"

	"github.com/besart951/go_infra_link/backend/internal/config"
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	applogger "github.com/besart951/go_infra_link/backend/pkg/logger"
)

type seedUserService interface {
	CreateWithPassword(ctx context.Context, user *domainUser.User, password string) error
}

func ensureSeedUser(cfg config.Config, log applogger.Logger, userService seedUserService, userEmailRepo domainUser.UserEmailRepository) error {
	if !cfg.SeedUserEnabled {
		return nil
	}
	if cfg.SeedUserEmail == "" || cfg.SeedUserPassword == "" {
		return nil
	}

	user, err := userEmailRepo.GetByEmail(context.Background(), cfg.SeedUserEmail)
	if err != nil {
		if !errors.Is(err, domain.ErrNotFound) {
			return err
		}

		user = &domainUser.User{
			FirstName: cfg.SeedUserFirstName,
			LastName:  cfg.SeedUserLastName,
			Email:     cfg.SeedUserEmail,
			IsActive:  true,
			Role:      domainUser.RoleSuperAdmin,
		}

		if err := userService.CreateWithPassword(context.Background(), user, cfg.SeedUserPassword); err != nil {
			return err
		}

		log.Info("Seeded initial superadmin user", "email", cfg.SeedUserEmail)
		return nil
	}

	log.Info(
		"Seed user already exists; skipping startup mutation",
		"email",
		cfg.SeedUserEmail,
		"user_id",
		user.ID,
	)
	return nil
}
