package facilitysql

import (
	"strings"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/repository/gormbase"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type spsControllerRepo struct {
	*gormbase.BaseRepository[*domainFacility.SPSController]
}

func NewSPSControllerRepository(db *gorm.DB) domainFacility.SPSControllerRepository {
	searchCallback := func(query *gorm.DB, search string) *gorm.DB {
		pattern := "%" + strings.ToLower(strings.TrimSpace(search)) + "%"
		return query.Where("LOWER(device_name) LIKE ? OR LOWER(ip_address) LIKE ?", pattern, pattern)
	}

	baseRepo := gormbase.NewBaseRepository[*domainFacility.SPSController](db, searchCallback)
	return &spsControllerRepo{BaseRepository: baseRepo}
}

func (r *spsControllerRepo) GetByIds(ids []uuid.UUID) ([]*domainFacility.SPSController, error) {
	return r.BaseRepository.GetByIds(ids)
}

func (r *spsControllerRepo) Create(entity *domainFacility.SPSController) error {
	return r.BaseRepository.Create(entity)
}

func (r *spsControllerRepo) Update(entity *domainFacility.SPSController) error {
	return r.BaseRepository.Update(entity)
}

func (r *spsControllerRepo) DeleteByIds(ids []uuid.UUID) error {
	return r.BaseRepository.DeleteByIds(ids)
}

func (r *spsControllerRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[domainFacility.SPSController], error) {
	result, err := r.BaseRepository.GetPaginatedList(params, 10)
	if err != nil {
		return nil, err
	}

	// Convert []*SPSController to []SPSController for the interface
	items := make([]domainFacility.SPSController, len(result.Items))
	for i, item := range result.Items {
		items[i] = *item
	}

	return &domain.PaginatedList[domainFacility.SPSController]{
		Items:      items,
		Total:      result.Total,
		Page:       result.Page,
		TotalPages: result.TotalPages,
	}, nil
}
