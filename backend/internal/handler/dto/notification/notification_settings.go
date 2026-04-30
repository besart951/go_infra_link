package notification

import (
	"time"

	"github.com/google/uuid"
)

type UpsertSMTPSettingsRequest struct {
	Enabled          *bool  `json:"enabled" binding:"required"`
	Host             string `json:"host" binding:"required"`
	Port             int    `json:"port" binding:"required,min=1,max=65535"`
	Username         string `json:"username"`
	Password         string `json:"password"`
	FromEmail        string `json:"from_email" binding:"required,email"`
	FromName         string `json:"from_name"`
	ReplyTo          string `json:"reply_to" binding:"omitempty,email"`
	Security         string `json:"security" binding:"required,oneof=none starttls tls"`
	AuthMode         string `json:"auth_mode" binding:"required,oneof=none plain"`
	AllowInsecureTLS *bool  `json:"allow_insecure_tls" binding:"required"`
}

type SendSMTPTestEmailRequest struct {
	To      string `json:"to" binding:"required,email"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

type UpsertUserNotificationPreferenceRequest struct {
	NotificationEmail string `json:"notification_email" binding:"omitempty,email"`
	Channel           string `json:"channel" binding:"required,oneof=email system both"`
	Frequency         string `json:"frequency" binding:"required,oneof=immediate hourly daily weekly"`
}

type VerifyUserNotificationEmailRequest struct {
	Code string `json:"code" binding:"required,len=6,numeric"`
}

type UpsertNotificationRuleRequest struct {
	Name             string      `json:"name" binding:"required"`
	Enabled          *bool       `json:"enabled" binding:"required"`
	EventKey         string      `json:"event_key" binding:"required"`
	ProjectID        *uuid.UUID  `json:"project_id"`
	ResourceType     string      `json:"resource_type"`
	ResourceID       *uuid.UUID  `json:"resource_id"`
	RecipientType    string      `json:"recipient_type" binding:"required,oneof=users team project_users project_role"`
	RecipientUserIDs []uuid.UUID `json:"recipient_user_ids"`
	RecipientTeamID  *uuid.UUID  `json:"recipient_team_id"`
	RecipientRole    string      `json:"recipient_role"`
}

type SMTPSettingsResponse struct {
	ID               uuid.UUID  `json:"id"`
	Provider         string     `json:"provider"`
	Enabled          bool       `json:"enabled"`
	Host             string     `json:"host"`
	Port             int        `json:"port"`
	Username         string     `json:"username"`
	HasPassword      bool       `json:"has_password"`
	FromEmail        string     `json:"from_email"`
	FromName         string     `json:"from_name"`
	ReplyTo          string     `json:"reply_to"`
	Security         string     `json:"security"`
	AuthMode         string     `json:"auth_mode"`
	AllowInsecureTLS bool       `json:"allow_insecure_tls"`
	UpdatedAt        time.Time  `json:"updated_at"`
	UpdatedByID      *uuid.UUID `json:"updated_by_id,omitempty"`
}

type UserNotificationPreferenceResponse struct {
	ID                          uuid.UUID  `json:"id,omitempty"`
	UserID                      uuid.UUID  `json:"user_id"`
	NotificationEmail           string     `json:"notification_email"`
	NotificationEmailVerifiedAt *time.Time `json:"notification_email_verified_at,omitempty"`
	EmailVerificationSentAt     *time.Time `json:"email_verification_sent_at,omitempty"`
	EmailVerificationExpiresAt  *time.Time `json:"email_verification_expires_at,omitempty"`
	Channel                     string     `json:"channel"`
	Frequency                   string     `json:"frequency"`
	CreatedAt                   time.Time  `json:"created_at,omitempty"`
	UpdatedAt                   time.Time  `json:"updated_at,omitempty"`
}

type SystemNotificationResponse struct {
	ID           uuid.UUID         `json:"id"`
	RecipientID  uuid.UUID         `json:"recipient_id"`
	ActorID      *uuid.UUID        `json:"actor_id,omitempty"`
	EventKey     string            `json:"event_key"`
	Title        string            `json:"title"`
	Body         string            `json:"body"`
	ResourceType string            `json:"resource_type"`
	ResourceID   *uuid.UUID        `json:"resource_id,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	ReadAt       *time.Time        `json:"read_at,omitempty"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

type SystemNotificationListResponse struct {
	Items       []SystemNotificationResponse `json:"items"`
	Total       int64                        `json:"total"`
	Page        int                          `json:"page"`
	TotalPages  int                          `json:"total_pages"`
	UnreadCount int64                        `json:"unread_count"`
}

type NotificationRuleResponse struct {
	ID               uuid.UUID   `json:"id"`
	Name             string      `json:"name"`
	Enabled          bool        `json:"enabled"`
	EventKey         string      `json:"event_key"`
	ProjectID        *uuid.UUID  `json:"project_id,omitempty"`
	ResourceType     string      `json:"resource_type"`
	ResourceID       *uuid.UUID  `json:"resource_id,omitempty"`
	RecipientType    string      `json:"recipient_type"`
	RecipientUserIDs []uuid.UUID `json:"recipient_user_ids,omitempty"`
	RecipientTeamID  *uuid.UUID  `json:"recipient_team_id,omitempty"`
	RecipientRole    string      `json:"recipient_role"`
	CreatedByID      *uuid.UUID  `json:"created_by_id,omitempty"`
	CreatedAt        time.Time   `json:"created_at"`
	UpdatedAt        time.Time   `json:"updated_at"`
}

type NotificationRuleListResponse struct {
	Items []NotificationRuleResponse `json:"items"`
}
