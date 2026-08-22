package facilitysql

import (
	"context"
	domainFieldDevice "github.com/besart951/go_infra_link/backend/internal/domain/facility/fielddevice"
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

func (r *fieldDeviceRepo) GetExportPage(ctx context.Context, filters domainFacility.FieldDeviceFilterParams, afterID uuid.UUID, limit int) ([]domainFacility.FieldDevice, error) {
	return newFieldDeviceQuery(r.db).ExportPage(ctx, filters, afterID, limit)
}

func (r *fieldDeviceRepo) GetCursorPage(ctx context.Context, query domainFacility.FieldDeviceCursorQuery) (*domainFacility.FieldDeviceCursorPage, error) {
	return newFieldDeviceQuery(r.db).CursorPage(ctx, query)
}

func (r *fieldDeviceRepo) GetExportControllerIDs(ctx context.Context, filters domainFacility.FieldDeviceFilterParams, search string) ([]uuid.UUID, error) {
	return newFieldDeviceQuery(r.db).ExportControllerIDs(ctx, filters, search)
}

func NewFieldDeviceRepository(db *gorm.DB) domainFieldDevice.FieldDeviceStore {
	return &fieldDeviceRepo{
		db: db,
	}
}

func (r *fieldDeviceRepo) GetByIds(ctx context.Context, ids []uuid.UUID) ([]*domainFacility.FieldDevice, error) {
	if len(ids) == 0 {
		return []*domainFacility.FieldDevice{}, nil
	}
	var records []*FieldDeviceRecord
	query := activeFieldDevices(r.db.WithContext(ctx).Model(&FieldDeviceRecord{}))
	if err := query.Where("field_devices.id IN ?", ids).Find(&records).Error; err != nil {
		return nil, err
	}
	items := toFieldDeviceDomains(records)
	if err := r.attachCanonicalSpecifications(ctx, items, ids); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *fieldDeviceRepo) attachCanonicalSpecifications(ctx context.Context, items []*domainFacility.FieldDevice, ids []uuid.UUID) error {
	var specifications []*domainFacility.Specification
	if err := r.db.WithContext(ctx).Where("field_device_id IN ?", ids).Find(&specifications).Error; err != nil {
		return err
	}
	byDeviceID := make(map[uuid.UUID]*domainFacility.Specification, len(specifications))
	for _, specification := range specifications {
		if specification != nil && specification.FieldDeviceID != nil {
			byDeviceID[*specification.FieldDeviceID] = specification
		}
	}
	for _, item := range items {
		attachCanonicalSpecification(item, byDeviceID[item.ID])
	}
	return nil
}

func attachCanonicalSpecification(item *domainFacility.FieldDevice, specification *domainFacility.Specification) {
	if item == nil || specification == nil {
		return
	}
	item.Specification = specification
	item.SpecificationID = &specification.ID
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
	mutation := activeMutation{
		target: activeMutationRequest{table: "field_devices", id: entity.ID, query: func(tx *gorm.DB) *gorm.DB {
			return activeFieldDevices(tx.Model(&FieldDeviceRecord{}))
		}},
		mutate: func(tx *gorm.DB) error { return updateFieldDeviceRecord(ctx, tx, entity) },
	}
	return runActiveMutation(ctx, r.db, mutation)
}

func updateFieldDeviceRecord(ctx context.Context, db *gorm.DB, entity *domainFacility.FieldDevice) error {
	expectedVersion := entity.Version
	entity.Base.TouchForUpdate(time.Now().UTC())
	record := toFieldDeviceRecord(entity)
	query := db.WithContext(ctx).Model(&FieldDeviceRecord{}).
		Where("id = ?", entity.ID)
	if expectedVersion > 0 {
		query = query.Where("version = ?", expectedVersion)
	}
	// Select("*") persists every scalar record field, including nil pointers.
	// That keeps new FieldDevice columns from being silently omitted by a second
	// hand-maintained update map. Primary and creation fields stay immutable and
	// associations remain outside this aggregate write.
	result := query.
		Select("*").
		Omit("id", "created_at", clause.Associations).
		Updates(record)
	if result.Error != nil {
		entity.Version = expectedVersion
		return result.Error
	}
	if expectedVersion > 0 && result.RowsAffected == 0 {
		entity.Version = expectedVersion
		return domain.ErrConflict
	}
	return nil
}

func (r *fieldDeviceRepo) DeleteByIds(ctx context.Context, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).Where("id IN ?", ids).Delete(&FieldDeviceRecord{}).Error
}

func (r *fieldDeviceRepo) DeleteAtVersion(ctx context.Context, command domainFacility.FieldDeviceDeleteCommand) error {
	if command.BaseVersion == nil {
		return domain.ErrInvalidArgument
	}
	query := activeFieldDevices(r.db.WithContext(ctx).Model(&FieldDeviceRecord{}))
	result := query.Where("field_devices.id = ? AND field_devices.version = ?", command.ID, *command.BaseVersion).Delete(&FieldDeviceRecord{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrConflict
	}
	return nil
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
