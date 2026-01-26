package facilitysql

import (
	"strings"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/repository/gormbase"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type fieldDeviceRepo struct {
	*gormbase.BaseRepository[*domainFacility.FieldDevice]
	db *gorm.DB
}

func NewFieldDeviceRepository(db *gorm.DB) domainFacility.FieldDeviceStore {
	searchCallback := func(query *gorm.DB, search string) *gorm.DB {
		pattern := "%" + strings.ToLower(strings.TrimSpace(search)) + "%"
		return query.Where("LOWER(bmk) LIKE ? OR LOWER(description) LIKE ?", pattern, pattern)
	}

	baseRepo := gormbase.NewBaseRepository[*domainFacility.FieldDevice](db, searchCallback)
	return &fieldDeviceRepo{
		BaseRepository: baseRepo,
		db:             db,
	}
}

func (r *fieldDeviceRepo) GetByIds(ids []uuid.UUID) ([]*domainFacility.FieldDevice, error) {
	// FieldDevice needs to preload Specification
	if len(ids) == 0 {
		return []*domainFacility.FieldDevice{}, nil
	}
	var items []*domainFacility.FieldDevice
	err := r.db.Where("deleted_at IS NULL").Where("id IN ?", ids).Preload("Specification").Find(&items).Error
	return items, err
}

func (r *fieldDeviceRepo) Create(entity *domainFacility.FieldDevice) error {
	return r.BaseRepository.Create(entity)
}

func (r *fieldDeviceRepo) Update(entity *domainFacility.FieldDevice) error {
	return r.BaseRepository.Update(entity)
}

func (r *fieldDeviceRepo) DeleteByIds(ids []uuid.UUID) error {
	return r.BaseRepository.DeleteByIds(ids)
}

func (r *fieldDeviceRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[domainFacility.FieldDevice], error) {
	result, err := r.BaseRepository.GetPaginatedList(params, 10)
	if err != nil {
		return nil, err
	}

	// Convert []*FieldDevice to []FieldDevice for the interface
	items := make([]domainFacility.FieldDevice, len(result.Items))
	for i, item := range result.Items {
		items[i] = *item
	}

	return &domain.PaginatedList[domainFacility.FieldDevice]{
		Items:      items,
		Total:      result.Total,
		Page:       result.Page,
		TotalPages: result.TotalPages,
	}, nil
}

func (r *fieldDeviceRepo) ExistsApparatNrConflict(spsControllerSystemTypeID uuid.UUID, systemPartID *uuid.UUID, apparatID uuid.UUID, apparatNr int, excludeID *uuid.UUID) (bool, error) {
	db := r.db.Model(&domainFacility.FieldDevice{}).
		Where("deleted_at IS NULL").
		Where("sps_controller_system_type_id = ?", spsControllerSystemTypeID).
		Where("apparat_id = ?", apparatID).
		Where("apparat_nr = ?", apparatNr)

	if systemPartID != nil {
		db = db.Where("system_part_id = ?", *systemPartID)
	} else {
		db = db.Where("system_part_id IS NULL")
	}

	if excludeID != nil {
		db = db.Where("id != ?", *excludeID)
	}

	var count int64
	err := db.Count(&count).Error
	return count > 0, err
}
