package userregistration

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	domainNotification "github.com/besart951/go_infra_link/backend/internal/domain/notification"
	"github.com/google/uuid"
)

type InvitationEmailBuilder interface {
	Build(userID uuid.UUID, email string, token string, expiresAt time.Time, nextAttemptAt time.Time) *domainNotification.EmailOutbox
}

type defaultInvitationEmailBuilder struct {
	appPublicURL string
}

func NewInvitationEmailBuilder(appPublicURL string) InvitationEmailBuilder {
	if strings.TrimSpace(appPublicURL) == "" {
		appPublicURL = defaultAppPublicURL
	}
	return &defaultInvitationEmailBuilder{appPublicURL: strings.TrimRight(strings.TrimSpace(appPublicURL), "/")}
}

func (b *defaultInvitationEmailBuilder) Build(userID uuid.UUID, email string, token string, expiresAt time.Time, nextAttemptAt time.Time) *domainNotification.EmailOutbox {
	link := b.appPublicURL + "/register/" + url.PathEscape(token)
	body := fmt.Sprintf(
		"Sie wurden zu Infra Link eingeladen.\n\nRegistrierung: %s\n\nDer Link ist bis %s UTC gueltig.\n\nDatenschutz: Ihre E-Mail-Adresse wird zur Kontoerstellung und Zustellung dieser Einladung verarbeitet. Weitere Profilangaben erfassen Sie im naechsten Schritt selbst.",
		link,
		expiresAt.UTC().Format(time.RFC3339),
	)
	return &domainNotification.EmailOutbox{
		RecipientID:    userID,
		RecipientEmail: email,
		EventKey:       registrationEventKey,
		Subject:        "Infra Link Registrierung",
		Body:           body,
		Frequency:      domainNotification.DeliveryFrequencyImmediate,
		Status:         domainNotification.EmailOutboxStatusPending,
		NextAttemptAt:  nextAttemptAt,
		Metadata: map[string]string{
			"purpose": "user_registration",
		},
	}
}
