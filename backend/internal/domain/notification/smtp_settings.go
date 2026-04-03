package notification

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/google/uuid"
)

type Provider string

const (
	ProviderSMTP Provider = "smtp"
)

type SecurityMode string

const (
	SecurityNone     SecurityMode = "none"
	SecuritySTARTTLS SecurityMode = "starttls"
	SecurityTLS      SecurityMode = "tls"
)

type AuthMode string

const (
	AuthModeNone  AuthMode = "none"
	AuthModePlain AuthMode = "plain"
)

type SMTPSettings struct {
	domain.Base
	Provider          Provider `gorm:"type:varchar(32);uniqueIndex;not null"`
	Enabled           bool     `gorm:"not null;default:true"`
	Host              string   `gorm:"not null"`
	Port              int      `gorm:"not null"`
	Username          string
	PasswordEncrypted string `gorm:"column:password_encrypted;type:text"`
	FromEmail         string `gorm:"not null"`
	FromName          string
	ReplyTo           string
	Security          SecurityMode `gorm:"type:varchar(16);not null;default:'starttls'"`
	AuthMode          AuthMode     `gorm:"type:varchar(16);not null;default:'plain'"`
	AllowInsecureTLS  bool         `gorm:"not null;default:false"`
	UpdatedByID       *uuid.UUID   `gorm:"type:uuid"`
}

func (s *SMTPSettings) GetBase() *domain.Base {
	return &s.Base
}

func (p Provider) Valid() bool {
	switch p {
	case ProviderSMTP:
		return true
	default:
		return false
	}
}

func (m SecurityMode) Valid() bool {
	switch m {
	case SecurityNone, SecuritySTARTTLS, SecurityTLS:
		return true
	default:
		return false
	}
}

func (m AuthMode) Valid() bool {
	switch m {
	case AuthModeNone, AuthModePlain:
		return true
	default:
		return false
	}
}
