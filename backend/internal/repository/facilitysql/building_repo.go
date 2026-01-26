package facilitysql

import (
	"strconv"
	"strings"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/repository/gormbase"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type buildingRepo struct {
	*gormbase.BaseRepository[*domainFacility.Building]
}

func NewBuildingRepository(db *gorm.DB) domainFacility.BuildingRepository {
	searchCallback := func(query *gorm.DB, search string) *gorm.DB {
		pattern := "%" + strings.ToLower(strings.TrimSpace(search)) + "%"
		if num, err := strconv.Atoi(strings.TrimSpace(search)); err == nil {
			return query.Where("LOWER(iws_code) LIKE ? OR building_group = ?", pattern, num)
		}
		return query.Where("LOWER(iws_code) LIKE ?", pattern)
	}

	baseRepo := gormbase.NewBaseRepository[*domainFacility.Building](db, searchCallback)
	return &buildingRepo{BaseRepository: baseRepo}
}

func (r *buildingRepo) GetByIds(ids []uuid.UUID) ([]*domainFacility.Building, error) {
	return r.BaseRepository.GetByIds(ids)
}

func (r *buildingRepo) Create(entity *domainFacility.Building) error {
	return r.BaseRepository.Create(entity)
}

func (r *buildingRepo) Update(entity *domainFacility.Building) error {
	return r.BaseRepository.Update(entity)
}

func (r *buildingRepo) DeleteByIds(ids []uuid.UUID) error {
	return r.BaseRepository.DeleteByIds(ids)
}

func (r *buildingRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[domainFacility.Building], error) {
	result, err := r.BaseRepository.GetPaginatedList(params, 10)
	if err != nil {
		return nil, err
	}

	// Convert []*Building to []Building for the interface
	items := make([]domainFacility.Building, len(result.Items))
	for i, item := range result.Items {
		items[i] = *item
	}

	return &domain.PaginatedList[domainFacility.Building]{
		Items:      items,
		Total:      result.Total,
		Page:       result.Page,
		TotalPages: result.TotalPages,
	}, nil
}
