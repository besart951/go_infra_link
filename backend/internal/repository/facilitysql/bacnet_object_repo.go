package facilitysql

import (
	"strings"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/repository/gormbase"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type bacnetObjectRepo struct {
	*gormbase.BaseRepository[*domainFacility.BacnetObject]
	db *gorm.DB
}

func NewBacnetObjectRepository(db *gorm.DB) domainFacility.BacnetObjectStore {
	searchCallback := func(query *gorm.DB, search string) *gorm.DB {
		pattern := "%" + strings.ToLower(strings.TrimSpace(search)) + "%"
		return query.Where("LOWER(text_fix) LIKE ?", pattern)
	}

	baseRepo := gormbase.NewBaseRepository[*domainFacility.BacnetObject](db, searchCallback)
	return &bacnetObjectRepo{
		BaseRepository: baseRepo,
		db:             db,
	}
}

func (r *bacnetObjectRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[domainFacility.BacnetObject], error) {
	result, err := r.BaseRepository.GetPaginatedList(params, 10)
	if err != nil {
		return nil, err
	}
	return gormbase.DerefPaginatedList(result), nil
}

func (r *bacnetObjectRepo) GetByFieldDeviceIDs(ids []uuid.UUID) ([]*domainFacility.BacnetObject, error) {
	if len(ids) == 0 {
		return []*domainFacility.BacnetObject{}, nil
	}
	var items []*domainFacility.BacnetObject
	err := r.db.Where("field_device_id IN ?", ids).Find(&items).Error
	return items, err
}

func (r *bacnetObjectRepo) DeleteByFieldDeviceIDs(ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	return r.db.
		Where("field_device_id IN ?", ids).
		Delete(&domainFacility.BacnetObject{}).Error
}
