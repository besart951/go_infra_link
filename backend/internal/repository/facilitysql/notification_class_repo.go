package facilitysql

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type notificationClassRepo struct {
	db *gorm.DB
}

func NewNotificationClassRepository(db *gorm.DB) domainFacility.NotificationClassRepository {
	return &notificationClassRepo{db: db}
}

func (r *notificationClassRepo) GetByIds(ids []uuid.UUID) ([]*domainFacility.NotificationClass, error) {
	var items []*domainFacility.NotificationClass
	err := r.db.Where("id IN ?", ids).Find(&items).Error
	return items, err
}

func (r *notificationClassRepo) Create(entity *domainFacility.NotificationClass) error {
	return r.db.Create(entity).Error
}

func (r *notificationClassRepo) Update(entity *domainFacility.NotificationClass) error {
	return r.db.Save(entity).Error
}

func (r *notificationClassRepo) DeleteByIds(ids []uuid.UUID) error {
	return r.db.Delete(&domainFacility.NotificationClass{}, ids).Error
}

func (r *notificationClassRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[domainFacility.NotificationClass], error) {
	var items []domainFacility.NotificationClass
	var total int64

	db := r.db.Model(&domainFacility.NotificationClass{})

	if params.Search != "" {
		db = db.Where("object_description ILIKE ? OR event_category ILIKE ? OR meaning ILIKE ?", 
			"%"+params.Search+"%", "%"+params.Search+"%", "%"+params.Search+"%")
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}

	offset := (params.Page - 1) * params.Limit
	if err := db.Limit(params.Limit).Offset(offset).Order("nc ASC").Find(&items).Error; err != nil {
		return nil, err
	}

	return &domain.PaginatedList[domainFacility.NotificationClass]{
		Items:      items,
		Total:      total,
		Page:       params.Page,
		TotalPages: domain.CalculateTotalPages(total, params.Limit),
	}, nil
}
