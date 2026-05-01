package notification

import (
	"context"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainTeam "github.com/besart951/go_infra_link/backend/internal/domain/team"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
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

type UpsertUserPreferenceInput struct {
	UserID            uuid.UUID
	NotificationEmail string
	Channel           DeliveryChannel
	Frequency         DeliveryFrequency
}

type SendUserPreferenceVerificationCodeInput struct {
	UserID uuid.UUID
}

type VerifyUserPreferenceEmailInput struct {
	UserID uuid.UUID
	Code   string
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

type DispatchNotificationInput struct {
	RecipientIDs []uuid.UUID
	ActorID      *uuid.UUID
	Locale       string
	EventKey     string
	Title        string
	Body         string
	ResourceType string
	ResourceID   *uuid.UUID
	Metadata     map[string]string
}

type DispatchEventInput struct {
	ActorID      *uuid.UUID
	Locale       string
	EventKey     string
	Title        string
	Body         string
	ProjectID    *uuid.UUID
	ResourceType string
	ResourceID   *uuid.UUID
	Metadata     map[string]string
}

type UpsertNotificationRuleInput struct {
	ID               uuid.UUID
	ActorID          uuid.UUID
	Name             string
	Enabled          bool
	EventKey         string
	ProjectID        *uuid.UUID
	ResourceType     string
	ResourceID       *uuid.UUID
	RecipientType    RuleRecipientType
	RecipientUserIDs []uuid.UUID
	RecipientTeamID  *uuid.UUID
	RecipientRole    domainUser.Role
}

type NotificationRuleFilter struct {
	EventKey  string
	ProjectID *uuid.UUID
	Enabled   *bool
}

type SMTPSettingsRepository interface {
	GetByProvider(ctx context.Context, provider Provider) (*SMTPSettings, error)
	Save(ctx context.Context, settings *SMTPSettings) error
}

type UserPreferenceRepository interface {
	GetByUserID(ctx context.Context, userID uuid.UUID) (*UserPreference, error)
	Save(ctx context.Context, preference *UserPreference) error
}

type SystemNotificationRepository interface {
	Create(ctx context.Context, notification *SystemNotification) error
	GetPaginatedListForUser(ctx context.Context, userID uuid.UUID, params domain.PaginationParams, unreadOnly bool) (*domain.PaginatedList[SystemNotification], error)
	CountUnreadForUser(ctx context.Context, userID uuid.UUID) (int64, error)
	MarkReadForUser(ctx context.Context, notificationID, userID uuid.UUID) (*SystemNotification, error)
	MarkAllReadForUser(ctx context.Context, userID uuid.UUID) error
	ToggleReadForUser(ctx context.Context, notificationID, userID uuid.UUID) (*SystemNotification, error)
	ToggleImportantForUser(ctx context.Context, notificationID, userID uuid.UUID) (*SystemNotification, error)
	DeleteForUser(ctx context.Context, notificationID, userID uuid.UUID) error
}

type EmailOutboxRepository interface {
	Create(ctx context.Context, item *EmailOutbox) error
	GetDue(ctx context.Context, now time.Time, limit int) ([]EmailOutbox, error)
	MarkSent(ctx context.Context, ids []uuid.UUID, sentAt time.Time) error
	MarkFailed(ctx context.Context, ids []uuid.UUID, attempts int, lastError string, nextAttemptAt time.Time) error
}

type NotificationRuleRepository interface {
	Create(ctx context.Context, rule *NotificationRule) error
	Update(ctx context.Context, rule *NotificationRule) error
	DeleteByID(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*NotificationRule, error)
	List(ctx context.Context, filter NotificationRuleFilter) ([]NotificationRule, error)
	ListMatching(ctx context.Context, eventKey string, projectID *uuid.UUID, resourceType string, resourceID *uuid.UUID) ([]NotificationRule, error)
}

type ProjectMembershipReader interface {
	ListUsers(ctx context.Context, projectID uuid.UUID) ([]domainUser.User, error)
}

type TeamMemberReader interface {
	ListByTeam(ctx context.Context, teamID uuid.UUID, params domain.PaginationParams) (*domain.PaginatedList[domainTeam.TeamMember], error)
}
