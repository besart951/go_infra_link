package app

import (
	"context"
	"errors"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/config"
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainNotification "github.com/besart951/go_infra_link/backend/internal/domain/notification"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	applogger "github.com/besart951/go_infra_link/backend/pkg/logger"
	"github.com/google/uuid"
)

type systemNotificationRepo interface {
	Create(ctx context.Context, notification *domainNotification.SystemNotification) error
	GetPaginatedListForUser(
		ctx context.Context,
		userID uuid.UUID,
		params domain.PaginationParams,
		unreadOnly bool,
	) (*domain.PaginatedList[domainNotification.SystemNotification], error)
}

type seedNotification = struct {
	eventKey string
	title    string
	body     string
	read     bool
	age      time.Duration
}

func ensureSeedSystemNotifications(
	cfg config.Config,
	log applogger.Logger,
	userEmailRepo domainUser.UserEmailRepository,
	systemRepo systemNotificationRepo,
) error {
	if !cfg.SeedDummyNotificationsEnabled {
		return nil
	}

	email := cfg.SeedDummyNotificationsEmail
	if email == "" {
		log.Info("Skipping dummy notification seeding; SEED_DUMMY_NOTIFICATIONS_EMAIL is empty")
		return nil
	}

	user, err := userEmailRepo.GetByEmail(context.Background(), email)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			log.Info("Skipping dummy notification seeding; seed user not found", "email", email)
			return nil
		}
		return err
	}

	existing, err := systemRepo.GetPaginatedListForUser(
		context.Background(),
		user.ID,
		domain.PaginationParams{Page: 1, Limit: 1},
		false,
	)
	if err != nil {
		return err
	}
	if existing.Total > 0 {
		log.Info("Skipping dummy notification seeding; user already has notifications", "email", email)
		return nil
	}

	now := time.Now().UTC()
	samples := []seedNotification{
		{
			eventKey: "project.created",
			title:    "Projekt wurde erstellt",
			body:     "Ihr neues Projekt \"Gebäude West\" ist erfolgreich angelegt.",
			read:     false,
			age:      12 * time.Hour,
		},
		{
			eventKey: "alarm.critical",
			title:    "Kritische Alarmmeldung",
			body:     "Temperaturüberschreitung am Sensor 4 wurde erkannt.",
			read:     false,
			age:      6 * time.Hour,
		},
		{
			eventKey: "maintenance.scheduled",
			title:    "Wartungsfenster geplant",
			body:     "Ein geplantes Wartungsfenster ist für morgen um 02:00 Uhr vorgesehen.",
			read:     true,
			age:      24 * time.Hour,
		},
		{
			eventKey: "team.member_invited",
			title:    "Team-Mitglied hinzugefügt",
			body:     "Neue Zugriffskette für Max Mustermann wurde freigegeben.",
			read:     true,
			age:      36 * time.Hour,
		},
		{
			eventKey: "report.generated",
			title:    "Wochenauswertung abgeschlossen",
			body:     "Die automatische Wochenauswertung ist fertig und bereit.",
			read:     false,
			age:      48 * time.Hour,
		},
		{
			eventKey: "system.update",
			title:    "System-Update verfügbar",
			body:     "Ein neues System-Update wurde für die Plattform freigegeben.",
			read:     true,
			age:      72 * time.Hour,
		},
	}

	for _, sample := range samples {
		createdAt := now.Add(-sample.age)
		item := &domainNotification.SystemNotification{
			Base: domain.Base{
				CreatedAt: createdAt,
				UpdatedAt: createdAt,
			},
			RecipientID: user.ID,
			EventKey:    sample.eventKey,
			Title:       sample.title,
			Body:        sample.body,
			Metadata:    map[string]string{"seed": "dummy"},
		}
		if sample.read {
			readAt := createdAt.Add(30 * time.Minute)
			item.ReadAt = &readAt
		}
		if err := systemRepo.Create(context.Background(), item); err != nil {
			return err
		}
	}

	log.Info("Seeded dummy system notifications", "email", email, "count", len(samples))
	return nil
}
