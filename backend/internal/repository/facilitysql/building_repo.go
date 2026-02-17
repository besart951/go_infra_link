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
	db *gorm.DB
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
	return &buildingRepo{BaseRepository: baseRepo, db: db}
}

func (r *buildingRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[domainFacility.Building], error) {
	result, err := r.BaseRepository.GetPaginatedList(params, 10)
	if err != nil {
		return nil, err
	}
	return gormbase.DerefPaginatedList(result), nil
}

func (r *buildingRepo) ExistsIWSCodeGroup(iwsCode string, buildingGroup int, excludeID *uuid.UUID) (bool, error) {
	query := r.db.Model(&domainFacility.Building{}).
		Where("deleted_at IS NULL").
		Where("LOWER(iws_code) = ?", strings.ToLower(strings.TrimSpace(iwsCode))).
		Where("building_group = ?", buildingGroup)

	if excludeID != nil {
		query = query.Where("id <> ?", *excludeID)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
