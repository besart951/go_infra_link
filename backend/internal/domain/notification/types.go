package notification

import (
	"context"

	"github.com/google/uuid"
)

type EmailMessage struct {
	To       []string
	Subject  string
	TextBody string
}

type UpsertSMTPSettingsInput struct {
	ActorID          uuid.UUID
	Enabled          bool
	Host             string
	Port             int
	Username         string
	Password         string
	FromEmail        string
	FromName         string
	ReplyTo          string
	Security         SecurityMode
	AuthMode         AuthMode
	AllowInsecureTLS bool
}

type SendTestEmailInput struct {
	To      string
	Subject string
	Body    string
}

type SendNotificationInput struct {
	To      []string
	Subject string
	Body    string
}

type SMTPSettingsRepository interface {
	GetByProvider(ctx context.Context, provider Provider) (*SMTPSettings, error)
	Save(ctx context.Context, settings *SMTPSettings) error
}
