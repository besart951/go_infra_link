package facilitysql

import (
	"strings"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/repository/gormbase"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type apparatRepo struct {
	*gormbase.BaseRepository[*domainFacility.Apparat]
}

func NewApparatRepository(db *gorm.DB) domainFacility.ApparatRepository {
	searchCallback := func(query *gorm.DB, search string) *gorm.DB {
		pattern := "%" + strings.ToLower(strings.TrimSpace(search)) + "%"
		return query.Where("LOWER(short_name) LIKE ? OR LOWER(name) LIKE ?", pattern, pattern)
	}

	baseRepo := gormbase.NewBaseRepository[*domainFacility.Apparat](db, searchCallback)
	return &apparatRepo{BaseRepository: baseRepo}
}

func (r *apparatRepo) GetByIds(ids []uuid.UUID) ([]*domainFacility.Apparat, error) {
	result, err := r.BaseRepository.GetByIds(ids)
	if err != nil {
		return nil, err
	}

	// Preload SystemParts for each apparat
	for _, apparat := range result {
		if err := r.DB().Model(apparat).Association("SystemParts").Find(&apparat.SystemParts); err != nil {
			return nil, err
		}
	}

	return result, nil
}

func (r *apparatRepo) Update(entity *domainFacility.Apparat) error {
	// Use GORM's Association API to replace SystemParts
	return r.DB().Transaction(func(tx *gorm.DB) error {
		// Update the entity itself
		if err := tx.Model(entity).Updates(entity).Error; err != nil {
			return err
		}

		// Replace SystemParts association
		if err := tx.Model(entity).Association("SystemParts").Replace(entity.SystemParts); err != nil {
			return err
		}

		return nil
	})
}

func (r *apparatRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[domainFacility.Apparat], error) {
	result, err := r.BaseRepository.GetPaginatedList(params, 10)
	if err != nil {
		return nil, err
	}

	// Preload SystemParts for each apparat
	for _, apparat := range result.Items {
		if err := r.DB().Model(apparat).Association("SystemParts").Find(&apparat.SystemParts); err != nil {
			return nil, err
		}
	}

	return gormbase.DerefPaginatedList(result), nil
}
