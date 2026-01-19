package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/repository"
	"github.com/google/uuid"
)

func (r *facilityRepo) GetNotificationClassByIds(ids []uuid.UUID) ([]*facility.NotificationClass, error) {
	var items []*facility.NotificationClass
	err := r.db.Where("id IN ?", ids).Find(&items).Error
	return items, err
}

func (r *facilityRepo) CreateNotificationClass(entity *facility.NotificationClass) error {
	return r.db.Create(entity).Error
}

func (r *facilityRepo) UpdateNotificationClass(entity *facility.NotificationClass) error {
	return r.db.Save(entity).Error
}

func (r *facilityRepo) DeleteNotificationClassByIds(ids []uuid.UUID) error {
	return r.db.Where("id IN ?", ids).Delete(&facility.NotificationClass{}).Error
}

func (r *facilityRepo) GetPaginatedNotificationClasses(params domain.PaginationParams) (*domain.PaginatedList[facility.NotificationClass], error) {
	return repository.Paginate[facility.NotificationClass](r.db, params, []string{"event_category", "meaning", "object_description"})
}
