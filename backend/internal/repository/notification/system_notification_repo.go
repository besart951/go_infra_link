package notification

import (
	"context"
	"errors"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainNotification "github.com/besart951/go_infra_link/backend/internal/domain/notification"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type systemNotificationRepo struct {
	db *gorm.DB
}

func NewSystemNotificationRepository(db *gorm.DB) domainNotification.SystemNotificationRepository {
	return &systemNotificationRepo{db: db}
}

func (r *systemNotificationRepo) Create(ctx context.Context, notification *domainNotification.SystemNotification) error {
	if err := notification.InitForCreate(time.Now().UTC()); err != nil {
		return err
	}
	return r.db.WithContext(ctx).Create(notification).Error
}

func (r *systemNotificationRepo) GetPaginatedListForUser(ctx context.Context, userID uuid.UUID, params domain.PaginationParams, unreadOnly bool) (*domain.PaginatedList[domainNotification.SystemNotification], error) {
	page, limit := domain.NormalizePagination(params.Page, params.Limit, 20)
	offset := (page - 1) * limit

	query := r.db.WithContext(ctx).
		Model(&domainNotification.SystemNotification{}).
		Where("recipient_id = ?", userID)
	if unreadOnly {
		query = query.Where("read_at IS NULL")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var items []domainNotification.SystemNotification
	if err := query.Order("is_important DESC, created_at DESC").Limit(limit).Offset(offset).Find(&items).Error; err != nil {
		return nil, err
	}

	return &domain.PaginatedList[domainNotification.SystemNotification]{
		Items:      items,
		Total:      total,
		Page:       page,
		TotalPages: domain.CalculateTotalPages(total, limit),
	}, nil
}

func (r *systemNotificationRepo) CountUnreadForUser(ctx context.Context, userID uuid.UUID) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).
		Model(&domainNotification.SystemNotification{}).
		Where("recipient_id = ? AND read_at IS NULL", userID).
		Count(&total).Error
	return total, err
}

func (r *systemNotificationRepo) MarkReadForUser(ctx context.Context, notificationID, userID uuid.UUID) (*domainNotification.SystemNotification, error) {
	var notification domainNotification.SystemNotification
	err := r.db.WithContext(ctx).
		Where("id = ? AND recipient_id = ?", notificationID, userID).
		First(&notification).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	notification.MarkRead(now)
	if err := r.db.WithContext(ctx).Model(&domainNotification.SystemNotification{}).
		Where("id = ?", notification.ID).
		Updates(map[string]any{
			"updated_at": notification.UpdatedAt,
			"read_at":    notification.ReadAt,
		}).Error; err != nil {
		return nil, err
	}

	return &notification, nil
}

func (r *systemNotificationRepo) MarkAllReadForUser(ctx context.Context, userID uuid.UUID) error {
	now := time.Now().UTC()
	return r.db.WithContext(ctx).Model(&domainNotification.SystemNotification{}).
		Where("recipient_id = ? AND read_at IS NULL", userID).
		Updates(map[string]any{
			"updated_at": now,
			"read_at":    now,
		}).Error
}

func (r *systemNotificationRepo) ToggleReadForUser(ctx context.Context, notificationID, userID uuid.UUID) (*domainNotification.SystemNotification, error) {
	var notification domainNotification.SystemNotification
	err := r.db.WithContext(ctx).
		Where("id = ? AND recipient_id = ?", notificationID, userID).
		First(&notification).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	notification.ToggleRead(now)
	if err := r.db.WithContext(ctx).Model(&domainNotification.SystemNotification{}).
		Where("id = ?", notification.ID).
		Updates(map[string]any{
			"updated_at": notification.UpdatedAt,
			"read_at":    notification.ReadAt,
		}).Error; err != nil {
		return nil, err
	}

	return &notification, nil
}

func (r *systemNotificationRepo) ToggleImportantForUser(ctx context.Context, notificationID, userID uuid.UUID) (*domainNotification.SystemNotification, error) {
	var notification domainNotification.SystemNotification
	err := r.db.WithContext(ctx).
		Where("id = ? AND recipient_id = ?", notificationID, userID).
		First(&notification).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	notification.ToggleImportant(now)
	if err := r.db.WithContext(ctx).Model(&domainNotification.SystemNotification{}).
		Where("id = ?", notification.ID).
		Updates(map[string]any{
			"updated_at":   notification.UpdatedAt,
			"is_important": notification.IsImportant,
		}).Error; err != nil {
		return nil, err
	}

	return &notification, nil
}

func (r *systemNotificationRepo) DeleteForUser(ctx context.Context, notificationID, userID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND recipient_id = ?", notificationID, userID).
		Delete(&domainNotification.SystemNotification{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}
