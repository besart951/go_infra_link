// Command e2e-seed creates the restricted account used only by the isolated
// browser-test environment. It intentionally refuses every other APP_ENV so
// the credentials can never be introduced by an operational deployment.
package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/config"
	"github.com/besart951/go_infra_link/backend/internal/db"
	domain "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	e2eEnvironment          = "e2e"
	plannerEmailKey         = "E2E_PLANNER_EMAIL"
	plannerPasswordKey      = "E2E_PLANNER_PASSWORD"
	collaboratorEmailKey    = "E2E_COLLABORATOR_EMAIL"
	collaboratorPasswordKey = "E2E_COLLABORATOR_PASSWORD"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	if !strings.EqualFold(cfg.AppEnv, e2eEnvironment) {
		log.Fatalf("%s may only run with APP_ENV=%s", "e2e-seed", e2eEnvironment)
	}

	email := strings.TrimSpace(os.Getenv(plannerEmailKey))
	password := os.Getenv(plannerPasswordKey)
	if email == "" || password == "" {
		log.Fatalf("%s and %s must be set", plannerEmailKey, plannerPasswordKey)
	}

	database, err := db.Connect(cfg.DBConfig)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}

	if err := ensureRestrictedPlanner(database, email, password, time.Now()); err != nil {
		log.Fatalf("seed restricted planner: %v", err)
	}

	collaboratorEmail := strings.TrimSpace(os.Getenv(collaboratorEmailKey))
	collaboratorPassword := os.Getenv(collaboratorPasswordKey)
	if collaboratorEmail == "" || collaboratorPassword == "" {
		log.Fatalf("%s and %s must be set", collaboratorEmailKey, collaboratorPasswordKey)
	}
	if err := ensureRealtimeCollaborator(database, collaboratorEmail, collaboratorPassword, time.Now()); err != nil {
		log.Fatalf("seed E2E collaborator: %v", err)
	}
	log.Printf("seeded restricted E2E planner: %s", domain.NormalizeEmail(email))
	log.Printf("seeded E2E realtime collaborator: %s", domain.NormalizeEmail(collaboratorEmail))
}

func ensureRestrictedPlanner(database *gorm.DB, email, password string, now time.Time) error {
	return ensureE2EUser(database, email, password, "E2E", "Planner", domain.RolePlaner, now)
}

func ensureRealtimeCollaborator(database *gorm.DB, email, password string, now time.Time) error {
	return ensureE2EUser(database, email, password, "E2E", "Collaborator", domain.RoleSuperAdmin, now)
}

func ensureE2EUser(database *gorm.DB, email, password, firstName, lastName string, role domain.Role, now time.Time) error {
	normalizedEmail := domain.NormalizeEmail(email)
	if normalizedEmail == "" {
		return errors.New("E2E user email is empty")
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash planner password: %w", err)
	}

	return database.Transaction(func(tx *gorm.DB) error {
		var planner domain.User
		err := tx.Where("email = ?", normalizedEmail).First(&planner).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			planner = domain.User{
				FirstName: firstName,
				LastName:  lastName,
				IsActive:  true,
				Role:      role,
				Password:  string(passwordHash),
			}
			planner.SetEmail(normalizedEmail)
			if err := planner.InitForCreate(now); err != nil {
				return fmt.Errorf("initialize planner: %w", err)
			}
			if err := tx.Create(&planner).Error; err != nil {
				return fmt.Errorf("create planner: %w", err)
			}
			return nil
		}
		if err != nil {
			return fmt.Errorf("find planner: %w", err)
		}

		planner.FirstName = firstName
		planner.LastName = lastName
		planner.IsActive = true
		planner.Role = role
		planner.Password = string(passwordHash)
		planner.DisabledAt = nil
		planner.DeletedAt = nil
		planner.DeletedByID = nil
		planner.RestoreUntil = nil
		planner.ScheduledPurgeAt = nil
		planner.AnonymizedAt = nil
		planner.DeletedEmailHash = nil
		planner.LockedUntil = nil
		planner.FailedLoginAttempts = 0
		planner.SetEmail(normalizedEmail)
		planner.TouchForUpdate(now)

		if err := tx.Save(&planner).Error; err != nil {
			return fmt.Errorf("update planner: %w", err)
		}
		return nil
	})
}
