package notification

import (
	"strings"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/google/uuid"
)

type RuleRecipientType string

const (
	RuleRecipientUsers        RuleRecipientType = "users"
	RuleRecipientTeam         RuleRecipientType = "team"
	RuleRecipientProjectUsers RuleRecipientType = "project_users"
	RuleRecipientProjectRole  RuleRecipientType = "project_role"
)

type NotificationRule struct {
	domain.Base
	Name             string            `gorm:"not null"`
	Enabled          bool              `gorm:"not null;default:true;index"`
	EventKey         string            `gorm:"type:varchar(128);not null;index"`
	ProjectID        *uuid.UUID        `gorm:"type:uuid;index"`
	ResourceType     string            `gorm:"type:varchar(64);index"`
	ResourceID       *uuid.UUID        `gorm:"type:uuid;index"`
	RecipientType    RuleRecipientType `gorm:"type:varchar(32);not null"`
	RecipientUserIDs []uuid.UUID       `gorm:"serializer:json;type:text"`
	RecipientTeamID  *uuid.UUID        `gorm:"type:uuid;index"`
	RecipientRole    domainUser.Role   `gorm:"type:varchar(50)"`
	CreatedByID      *uuid.UUID        `gorm:"type:uuid"`
}

func (r *NotificationRule) GetBase() *domain.Base {
	return &r.Base
}

func (t RuleRecipientType) Valid() bool {
	switch t {
	case RuleRecipientUsers, RuleRecipientTeam, RuleRecipientProjectUsers, RuleRecipientProjectRole:
		return true
	default:
		return false
	}
}

func NormalizeRuleRecipientType(value RuleRecipientType) RuleRecipientType {
	return RuleRecipientType(strings.ToLower(strings.TrimSpace(string(value))))
}
