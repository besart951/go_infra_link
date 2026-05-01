package facilitysql

import (
	"context"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/repository/gormbase"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type fieldDeviceRepo struct {
	db *gorm.DB
}

func NewFieldDeviceRepository(db *gorm.DB) domainFacility.FieldDeviceStore {
	return &fieldDeviceRepo{
		db: db,
	}
}

func (r *fieldDeviceRepo) GetByIds(ctx context.Context, ids []uuid.UUID) ([]*domainFacility.FieldDevice, error) {
	// FieldDevice needs to preload Specification
	if len(ids) == 0 {
		return []*domainFacility.FieldDevice{}, nil
	}
	var records []*FieldDeviceRecord
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Preload("Specification").Find(&records).Error
	return toFieldDeviceDomains(records), err
}

func (r *fieldDeviceRepo) Create(ctx context.Context, entity *domainFacility.FieldDevice) error {
	if err := entity.Base.InitForCreate(time.Now().UTC()); err != nil {
		return err
	}

	return r.db.WithContext(ctx).
		Omit(clause.Associations).
		Create(toFieldDeviceRecord(entity)).Error
}

func (r *fieldDeviceRepo) BulkCreate(ctx context.Context, entities []*domainFacility.FieldDevice, batchSize int) error {
	if len(entities) == 0 {
		return nil
	}

	now := time.Now().UTC()
	records := make([]*FieldDeviceRecord, len(entities))
	for i, entity := range entities {
		if err := entity.Base.InitForCreate(now); err != nil {
			return err
		}
		records[i] = toFieldDeviceRecord(entity)
	}

	if batchSize <= 0 {
		batchSize = gormbase.DefaultBatchSize
	}

	return r.db.WithContext(ctx).
		Omit(clause.Associations).
		CreateInBatches(records, batchSize).Error
}

func (r *fieldDeviceRepo) Update(ctx context.Context, entity *domainFacility.FieldDevice) error {
	entity.Base.TouchForUpdate(time.Now().UTC())
	return r.db.WithContext(ctx).Model(&FieldDeviceRecord{}).
		Where("id = ?", entity.ID).
		Updates(map[string]any{
			"updated_at":                    entity.UpdatedAt,
			"bmk":                           entity.BMK,
			"description":                   entity.Description,
			"apparat_nr":                    entity.ApparatNr,
			"text_individuell":              entity.TextIndividuell,
			"sps_controller_system_type_id": entity.SPSControllerSystemTypeID,
			"system_part_id":                entity.SystemPartID,
			"specification_id":              entity.SpecificationID,
			"apparat_id":                    entity.ApparatID,
		}).Error
}

func (r *fieldDeviceRepo) DeleteByIds(ctx context.Context, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).Where("id IN ?", ids).Delete(&FieldDeviceRecord{}).Error
}

func (r *fieldDeviceRepo) DeleteBySPSControllerSystemTypeIDs(ctx context.Context, systemTypeIDs []uuid.UUID) error {
	if len(systemTypeIDs) == 0 {
		return nil
	}

	for _, chunk := range uuidFilterChunks(systemTypeIDs, uuidFilterChunkSize) {
		if err := r.db.WithContext(ctx).
			Where("sps_controller_system_type_id IN ?", chunk).
			Delete(&FieldDeviceRecord{}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *fieldDeviceRepo) GetPaginatedList(ctx context.Context, params domain.PaginationParams) (*domain.PaginatedList[domainFacility.FieldDevice], error) {
	return newFieldDeviceQuery(r.db).List(ctx, params)
}

func (r *fieldDeviceRepo) GetIDsBySPSControllerSystemTypeIDs(ctx context.Context, ids []uuid.UUID) ([]uuid.UUID, error) {
	if len(ids) == 0 {
		return []uuid.UUID{}, nil
	}
	var out []uuid.UUID
	err := r.db.WithContext(ctx).Model(&FieldDeviceRecord{}).
		Where("sps_controller_system_type_id IN ?", ids).
		Pluck("id", &out).Error
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (r *fieldDeviceRepo) ExistsApparatNrConflict(ctx context.Context, spsControllerSystemTypeID uuid.UUID, systemPartID uuid.UUID, apparatID uuid.UUID, apparatNr int, excludeIDs []uuid.UUID) (bool, error) {
	db := r.db.WithContext(ctx).Model(&FieldDeviceRecord{}).
		Where("sps_controller_system_type_id = ?", spsControllerSystemTypeID).
		Where("system_part_id = ?", systemPartID).
		Where("apparat_id = ?", apparatID).
		Where("apparat_nr = ?", apparatNr)

	if len(excludeIDs) > 0 {
		db = db.Where("id NOT IN ?", excludeIDs)
	}

	var count int64
	err := db.Count(&count).Error
	return count > 0, err
}

func (r *fieldDeviceRepo) GetUsedApparatNumbers(ctx context.Context, spsControllerSystemTypeID uuid.UUID, systemPartID uuid.UUID, apparatID uuid.UUID) ([]int, error) {
	query := r.db.WithContext(ctx).Model(&FieldDeviceRecord{}).
		Where("sps_controller_system_type_id = ?", spsControllerSystemTypeID).
		Where("system_part_id = ?", systemPartID).
		Where("apparat_id = ?", apparatID)

	var nums []int
	if err := query.Pluck("apparat_nr", &nums).Error; err != nil {
		return nil, err
	}
	return nums, nil
}

func (r *fieldDeviceRepo) GetPaginatedListWithFilters(ctx context.Context, params domain.PaginationParams, filters domainFacility.FieldDeviceFilterParams) (*domain.PaginatedList[domainFacility.FieldDevice], error) {
	return newFieldDeviceQuery(r.db).ListWithFilters(ctx, params, filters)
}
