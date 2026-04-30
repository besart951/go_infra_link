package facilitysql

import (
	"context"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/repository/gormbase"
	"github.com/besart951/go_infra_link/backend/internal/repository/searchspec"
	"gorm.io/gorm"
)

type notificationClassRepo struct {
	*gormbase.BaseRepository[*domainFacility.NotificationClass]
}

func NewNotificationClassRepository(db *gorm.DB) domainFacility.NotificationClassRepository {
	baseRepo := gormbase.NewBaseRepository(db,
		gormbase.TrigramSearchCallback[*domainFacility.NotificationClass](searchspec.NotificationClasses.SearchColumns("")...),
	)
	return &notificationClassRepo{BaseRepository: baseRepo}
}

func (r *notificationClassRepo) GetPaginatedList(ctx context.Context, params domain.PaginationParams) (*domain.PaginatedList[domainFacility.NotificationClass], error) {
	result, err := r.BaseRepository.GetPaginatedList(ctx, params, 10)
	if err != nil {
		return nil, err
	}
	return gormbase.DerefPaginatedList(result), nil
}
