package facilitysql

import (
	"strings"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/repository/gormbase"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type systemPartRepo struct {
	*gormbase.BaseRepository[*domainFacility.SystemPart]
}

func NewSystemPartRepository(db *gorm.DB) domainFacility.SystemPartRepository {
	searchCallback := func(query *gorm.DB, search string) *gorm.DB {
		pattern := "%" + strings.ToLower(strings.TrimSpace(search)) + "%"
		return query.Where("LOWER(short_name) LIKE ? OR LOWER(name) LIKE ?", pattern, pattern)
	}

	baseRepo := gormbase.NewBaseRepository[*domainFacility.SystemPart](db, searchCallback)
	return &systemPartRepo{BaseRepository: baseRepo}
}

func (r *systemPartRepo) GetByIds(ids []uuid.UUID) ([]*domainFacility.SystemPart, error) {
	result, err := r.BaseRepository.GetByIds(ids)
	if err != nil {
		return nil, err
	}

	// Preload Apparats for each system part (many2many)
	for _, systemPart := range result {
		if err := r.DB().Model(systemPart).Association("Apparats").Find(&systemPart.Apparats); err != nil {
			return nil, err
		}
	}

	return result, nil
}

func (r *systemPartRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[domainFacility.SystemPart], error) {
	result, err := r.BaseRepository.GetPaginatedList(params, 10)
	if err != nil {
		return nil, err
	}
	return gormbase.DerefPaginatedList(result), nil
}
