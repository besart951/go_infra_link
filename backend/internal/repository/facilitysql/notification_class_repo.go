package facilitysql

import (
	"strings"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/repository/gormbase"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type notificationClassRepo struct {
	*gormbase.BaseRepository[*domainFacility.NotificationClass]
}

func NewNotificationClassRepository(db *gorm.DB) domainFacility.NotificationClassRepository {
	searchCallback := func(query *gorm.DB, search string) *gorm.DB {
		pattern := "%" + strings.TrimSpace(search) + "%"
		return query.Where("object_description ILIKE ? OR event_category ILIKE ? OR meaning ILIKE ?",
			pattern, pattern, pattern)
	}

	baseRepo := gormbase.NewBaseRepository[*domainFacility.NotificationClass](db, searchCallback)
	return &notificationClassRepo{BaseRepository: baseRepo}
}

func (r *notificationClassRepo) GetByIds(ids []uuid.UUID) ([]*domainFacility.NotificationClass, error) {
	return r.BaseRepository.GetByIds(ids)
}

func (r *notificationClassRepo) Create(entity *domainFacility.NotificationClass) error {
	return r.BaseRepository.Create(entity)
}

func (r *notificationClassRepo) Update(entity *domainFacility.NotificationClass) error {
	return r.BaseRepository.Update(entity)
}

func (r *notificationClassRepo) DeleteByIds(ids []uuid.UUID) error {
	return r.BaseRepository.DeleteByIds(ids)
}

func (r *notificationClassRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[domainFacility.NotificationClass], error) {
	result, err := r.BaseRepository.GetPaginatedList(params, 10)
	if err != nil {
		return nil, err
	}

	// Convert []*NotificationClass to []NotificationClass for the interface
	items := make([]domainFacility.NotificationClass, len(result.Items))
	for i, item := range result.Items {
		items[i] = *item
	}

	return &domain.PaginatedList[domainFacility.NotificationClass]{
		Items:      items,
		Total:      result.Total,
		Page:       result.Page,
		TotalPages: result.TotalPages,
	}, nil
}
