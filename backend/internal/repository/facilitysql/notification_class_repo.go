package facilitysql

import (
	"strings"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/repository/gormbase"
	"gorm.io/gorm"
)

type notificationClassRepo struct {
	*gormbase.BaseRepository[*domainFacility.NotificationClass]
}

func NewNotificationClassRepository(db *gorm.DB) domainFacility.NotificationClassRepository {
	searchCallback := func(query *gorm.DB, search string) *gorm.DB {
		pattern := "%" + strings.ToLower(strings.TrimSpace(search)) + "%"
		return query.Where("LOWER(object_description) LIKE ? OR LOWER(event_category) LIKE ? OR LOWER(meaning) LIKE ?",
			pattern, pattern, pattern)
	}

	baseRepo := gormbase.NewBaseRepository[*domainFacility.NotificationClass](db, searchCallback)
	return &notificationClassRepo{BaseRepository: baseRepo}
}

func (r *notificationClassRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[domainFacility.NotificationClass], error) {
	result, err := r.BaseRepository.GetPaginatedList(params, 10)
	if err != nil {
		return nil, err
	}
	return gormbase.DerefPaginatedList(result), nil
}
