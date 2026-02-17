package facilitysql

import (
	"strings"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/repository/gormbase"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type specificationRepo struct {
	*gormbase.BaseRepository[*domainFacility.Specification]
	db *gorm.DB
}

func NewSpecificationRepository(db *gorm.DB) domainFacility.SpecificationStore {
	searchCallback := func(query *gorm.DB, search string) *gorm.DB {
		pattern := "%" + strings.ToLower(strings.TrimSpace(search)) + "%"
		return query.Where("LOWER(specification_supplier) LIKE ? OR LOWER(specification_brand) LIKE ? OR LOWER(specification_type) LIKE ?", pattern, pattern, pattern)
	}

	baseRepo := gormbase.NewBaseRepository[*domainFacility.Specification](db, searchCallback)
	return &specificationRepo{
		BaseRepository: baseRepo,
		db:             db,
	}
}

func (r *specificationRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[domainFacility.Specification], error) {
	result, err := r.BaseRepository.GetPaginatedList(params, 10)
	if err != nil {
		return nil, err
	}
	return gormbase.DerefPaginatedList(result), nil
}

func (r *specificationRepo) GetByFieldDeviceIDs(fieldDeviceIDs []uuid.UUID) ([]*domainFacility.Specification, error) {
	if len(fieldDeviceIDs) == 0 {
		return []*domainFacility.Specification{}, nil
	}
	var items []*domainFacility.Specification
	err := r.db.Where("deleted_at IS NULL").Where("field_device_id IN ?", fieldDeviceIDs).Find(&items).Error
	return items, err
}

func (r *specificationRepo) SoftDeleteByFieldDeviceIDs(fieldDeviceIDs []uuid.UUID) error {
	if len(fieldDeviceIDs) == 0 {
		return nil
	}
	now := time.Now().UTC()
	return r.db.Model(&domainFacility.Specification{}).
		Where("field_device_id IN ?", fieldDeviceIDs).
		Updates(map[string]any{"deleted_at": now, "updated_at": now}).Error
}
