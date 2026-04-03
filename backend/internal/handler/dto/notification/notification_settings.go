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
