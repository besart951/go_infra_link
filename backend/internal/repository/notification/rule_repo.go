package notification

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainNotification "github.com/besart951/go_infra_link/backend/internal/domain/notification"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ruleRepo struct {
	db *gorm.DB
}

func NewNotificationRuleRepository(db *gorm.DB) domainNotification.NotificationRuleRepository {
	return &ruleRepo{db: db}
}

func (r *ruleRepo) Create(ctx context.Context, rule *domainNotification.NotificationRule) error {
	if err := rule.InitForCreate(time.Now().UTC()); err != nil {
		return err
	}
	return r.db.WithContext(ctx).Create(rule).Error
}

func (r *ruleRepo) Update(ctx context.Context, rule *domainNotification.NotificationRule) error {
	rule.TouchForUpdate(time.Now().UTC())
	return r.db.WithContext(ctx).Model(&domainNotification.NotificationRule{}).
		Where("id = ?", rule.ID).
		Updates(map[string]any{
			"updated_at":         rule.UpdatedAt,
			"name":               rule.Name,
			"enabled":            rule.Enabled,
			"event_key":          rule.EventKey,
			"project_id":         rule.ProjectID,
			"resource_type":      rule.ResourceType,
			"resource_id":        rule.ResourceID,
			"recipient_type":     rule.RecipientType,
			"recipient_user_ids": rule.RecipientUserIDs,
			"recipient_team_id":  rule.RecipientTeamID,
			"recipient_role":     rule.RecipientRole,
		}).Error
}

func (r *ruleRepo) DeleteByID(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return domain.ErrInvalidArgument
	}
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&domainNotification.NotificationRule{}).Error
}

func (r *ruleRepo) GetByID(ctx context.Context, id uuid.UUID) (*domainNotification.NotificationRule, error) {
	var rule domainNotification.NotificationRule
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&rule).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *ruleRepo) List(ctx context.Context, filter domainNotification.NotificationRuleFilter) ([]domainNotification.NotificationRule, error) {
	query := r.db.WithContext(ctx).Model(&domainNotification.NotificationRule{})
	if strings.TrimSpace(filter.EventKey) != "" {
		query = query.Where("event_key = ?", strings.TrimSpace(filter.EventKey))
	}
	if filter.ProjectID != nil {
		query = query.Where("project_id = ?", *filter.ProjectID)
	}
	if filter.Enabled != nil {
		query = query.Where("enabled = ?", *filter.Enabled)
	}

	var rules []domainNotification.NotificationRule
	err := query.Order("created_at DESC").Find(&rules).Error
	return rules, err
}

func (r *ruleRepo) ListMatching(ctx context.Context, eventKey string, projectID *uuid.UUID, resourceType string, resourceID *uuid.UUID) ([]domainNotification.NotificationRule, error) {
	query := r.db.WithContext(ctx).Model(&domainNotification.NotificationRule{}).
		Where("enabled = ? AND event_key = ?", true, strings.TrimSpace(eventKey))

	if projectID != nil {
		query = query.Where("(project_id IS NULL OR project_id = ?)", *projectID)
	} else {
		query = query.Where("project_id IS NULL")
	}

	resourceType = strings.TrimSpace(resourceType)
	if resourceType != "" {
		query = query.Where("(resource_type = '' OR resource_type = ?)", resourceType)
	} else {
		query = query.Where("resource_type = ''")
	}

	if resourceID != nil {
		query = query.Where("(resource_id IS NULL OR resource_id = ?)", *resourceID)
	} else {
		query = query.Where("resource_id IS NULL")
	}

	var rules []domainNotification.NotificationRule
	err := query.Order("created_at ASC").Find(&rules).Error
	return rules, err
}
