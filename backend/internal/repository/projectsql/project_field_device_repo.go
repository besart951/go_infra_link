package projectsql

import (
	"context"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type projectFieldDeviceRepo struct {
	db *gorm.DB
}

func NewProjectFieldDeviceRepository(db *gorm.DB) project.ProjectFieldDeviceRepository {
	return &projectFieldDeviceRepo{
		db: db,
	}
}

func (r *projectFieldDeviceRepo) GetByIds(ctx context.Context, ids []uuid.UUID) ([]*project.ProjectFieldDevice, error) {
	if len(ids) == 0 {
		return []*project.ProjectFieldDevice{}, nil
	}

	var records []*ProjectFieldDeviceRecord
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&records).Error
	return toProjectFieldDeviceDomains(records), err
}

func (r *projectFieldDeviceRepo) Create(ctx context.Context, entity *project.ProjectFieldDevice) error {
	if err := entity.Base.InitForCreate(time.Now().UTC()); err != nil {
		return err
	}
	if err := mapWriteError(r.db.WithContext(ctx).Create(toProjectFieldDeviceRecord(entity)).Error); err != nil {
		return err
	}
	return r.AddAssignmentSource(
		ctx,
		entity.ProjectID,
		[]uuid.UUID{entity.FieldDeviceID},
		project.ExplicitAssignmentSource(),
	)
}

func (r *projectFieldDeviceRepo) BulkCreate(ctx context.Context, entities []*project.ProjectFieldDevice, batchSize int) error {
	if len(entities) == 0 {
		return nil
	}
	if batchSize <= 0 {
		batchSize = 200
	}

	now := time.Now().UTC()
	records := make([]ProjectFieldDeviceRecord, 0, len(entities))
	for _, entity := range entities {
		if entity == nil {
			continue
		}
		if err := entity.Base.InitForCreate(now); err != nil {
			return err
		}
		records = append(records, *toProjectFieldDeviceRecord(entity))
	}
	if len(records) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		CreateInBatches(records, batchSize).Error
}

func (r *projectFieldDeviceRepo) BulkCreateByFieldDeviceIDs(ctx context.Context, projectID uuid.UUID, fieldDeviceIDs []uuid.UUID) error {
	_, err := r.BulkCreateByFieldDeviceIDsReturningIDs(ctx, projectID, fieldDeviceIDs)
	return err
}

// BulkCreateByFieldDeviceIDsReturningIDs returns only rows inserted by this
// statement so transactional history excludes pre-existing project links.
func (r *projectFieldDeviceRepo) BulkCreateByFieldDeviceIDsReturningIDs(
	ctx context.Context,
	projectID uuid.UUID,
	fieldDeviceIDs []uuid.UUID,
) ([]uuid.UUID, error) {
	return r.BulkCreateByFieldDeviceIDsWithSourceReturningIDs(
		ctx,
		projectID,
		fieldDeviceIDs,
		project.ExplicitAssignmentSource(),
	)
}

func (r *projectFieldDeviceRepo) BulkCreateByFieldDeviceIDsWithSourceReturningIDs(
	ctx context.Context,
	projectID uuid.UUID,
	fieldDeviceIDs []uuid.UUID,
	source project.AssignmentSource,
) ([]uuid.UUID, error) {
	if len(fieldDeviceIDs) == 0 {
		return nil, nil
	}
	if err := source.Validate(); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	seed := projectLinkSeed("project-field-device", projectID)
	const statement = `
		INSERT INTO project_field_devices (id, created_at, updated_at, project_id, field_device_id)
		SELECT ` + deterministicProjectLinkIDExpression + `, ?, ?, ?, field_devices.id
		FROM field_devices
		CROSS JOIN LATERAL (SELECT md5(? || field_devices.id::text) AS value) AS link_hash
		WHERE field_devices.id IN ?
		ON CONFLICT (project_id, field_device_id) DO NOTHING
		RETURNING id
	`

	insertedIDs := make([]uuid.UUID, 0, len(fieldDeviceIDs))
	for _, chunk := range uuidChunks(fieldDeviceIDs, projectLinkIDFilterChunkSize) {
		var inserted []struct {
			ID uuid.UUID `gorm:"column:id"`
		}
		if err := r.db.WithContext(ctx).
			Raw(statement, now, now, projectID, seed, chunk).
			Scan(&inserted).Error; err != nil {
			return nil, err
		}
		for _, row := range inserted {
			insertedIDs = append(insertedIDs, row.ID)
		}
	}
	if err := r.AddAssignmentSource(ctx, projectID, fieldDeviceIDs, source); err != nil {
		return nil, err
	}
	return insertedIDs, nil
}

func (r *projectFieldDeviceRepo) BulkCreateBySPSControllerSystemTypeIDs(ctx context.Context, projectID uuid.UUID, systemTypeIDs []uuid.UUID) error {
	_, err := r.BulkCreateBySPSControllerSystemTypeIDsReturningIDs(ctx, projectID, systemTypeIDs)
	return err
}

// BulkCreateBySPSControllerSystemTypeIDsReturningIDs returns the exact
// successful insert set for history correlation.
func (r *projectFieldDeviceRepo) BulkCreateBySPSControllerSystemTypeIDsReturningIDs(
	ctx context.Context,
	projectID uuid.UUID,
	systemTypeIDs []uuid.UUID,
) ([]uuid.UUID, error) {
	return r.BulkCreateBySPSControllerSystemTypeIDsWithSourceReturningIDs(
		ctx,
		projectID,
		systemTypeIDs,
		project.ExplicitAssignmentSource(),
	)
}

func (r *projectFieldDeviceRepo) BulkCreateBySPSControllerSystemTypeIDsWithSourceReturningIDs(
	ctx context.Context,
	projectID uuid.UUID,
	systemTypeIDs []uuid.UUID,
	source project.AssignmentSource,
) ([]uuid.UUID, error) {
	if len(systemTypeIDs) == 0 {
		return nil, nil
	}
	if err := source.Validate(); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	seed := projectLinkSeed("project-field-device", projectID)
	const statement = `
		INSERT INTO project_field_devices (id, created_at, updated_at, project_id, field_device_id)
		SELECT ` + deterministicProjectLinkIDExpression + `, ?, ?, ?, field_devices.id
		FROM field_devices
		CROSS JOIN LATERAL (SELECT md5(? || field_devices.id::text) AS value) AS link_hash
		WHERE field_devices.sps_controller_system_type_id IN ?
		ON CONFLICT (project_id, field_device_id) DO NOTHING
		RETURNING id
	`

	insertedIDs := make([]uuid.UUID, 0)
	for _, chunk := range uuidChunks(systemTypeIDs, projectFieldDeviceSystemTypeFilterChunkSize) {
		var inserted []struct {
			ID uuid.UUID `gorm:"column:id"`
		}
		if err := r.db.WithContext(ctx).
			Raw(statement, now, now, projectID, seed, chunk).
			Scan(&inserted).Error; err != nil {
			return nil, err
		}
		for _, row := range inserted {
			insertedIDs = append(insertedIDs, row.ID)
		}
	}
	fieldDeviceIDs := make([]uuid.UUID, 0)
	for _, chunk := range uuidChunks(systemTypeIDs, projectFieldDeviceSystemTypeFilterChunkSize) {
		var chunkIDs []uuid.UUID
		if err := r.db.WithContext(ctx).
			Table("field_devices").
			Where("sps_controller_system_type_id IN ?", chunk).
			Pluck("id", &chunkIDs).Error; err != nil {
			return nil, err
		}
		fieldDeviceIDs = append(fieldDeviceIDs, chunkIDs...)
	}
	if err := r.AddAssignmentSource(ctx, projectID, fieldDeviceIDs, source); err != nil {
		return nil, err
	}
	return insertedIDs, nil
}

func (r *projectFieldDeviceRepo) Update(ctx context.Context, entity *project.ProjectFieldDevice) error {
	if entity == nil || entity.ID == uuid.Nil || entity.Revision == 0 {
		return domain.ErrInvalidArgument
	}
	expectedRevision := entity.Revision
	entity.Base.TouchForUpdate(time.Now().UTC())
	result := r.db.WithContext(ctx).Model(&ProjectFieldDeviceRecord{}).
		Where("id = ? AND revision = ?", entity.ID, expectedRevision).
		Updates(map[string]any{
			"updated_at":      entity.UpdatedAt,
			"revision":        expectedRevision + 1,
			"project_id":      entity.ProjectID,
			"field_device_id": entity.FieldDeviceID,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return projectLinkRevisionConflict(
			ctx,
			r.db,
			&ProjectFieldDeviceRecord{},
			entity.ID,
			expectedRevision,
		)
	}
	entity.Revision = expectedRevision + 1
	return nil
}

func (r *projectFieldDeviceRepo) DeleteByIds(ctx context.Context, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Where("id IN ?", ids).Delete(&ProjectFieldDeviceRecord{}).Error
}

func (r *projectFieldDeviceRepo) GetPaginatedList(ctx context.Context, params domain.PaginationParams) (*domain.PaginatedList[project.ProjectFieldDevice], error) {
	page, limit := domain.NormalizePagination(params.Page, params.Limit, 10)
	offset := (page - 1) * limit

	query := r.db.WithContext(ctx).Model(&ProjectFieldDeviceRecord{})

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var records []ProjectFieldDeviceRecord
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&records).Error; err != nil {
		return nil, err
	}

	return &domain.PaginatedList[project.ProjectFieldDevice]{
		Items:      projectFieldDeviceDomainValues(records),
		Total:      total,
		Page:       page,
		TotalPages: domain.CalculateTotalPages(total, limit),
	}, nil
}

func (r *projectFieldDeviceRepo) GetPaginatedListByProjectID(ctx context.Context, projectID uuid.UUID, params domain.PaginationParams) (*domain.PaginatedList[project.ProjectFieldDevice], error) {
	page, limit := domain.NormalizePagination(params.Page, params.Limit, 10)
	offset := (page - 1) * limit

	query := r.db.WithContext(ctx).Model(&ProjectFieldDeviceRecord{}).
		Where("project_id = ?", projectID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var records []ProjectFieldDeviceRecord
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&records).Error; err != nil {
		return nil, err
	}

	return &domain.PaginatedList[project.ProjectFieldDevice]{
		Items:      projectFieldDeviceDomainValues(records),
		Total:      total,
		Page:       page,
		TotalPages: domain.CalculateTotalPages(total, limit),
	}, nil
}

func (r *projectFieldDeviceRepo) GetByProjectID(ctx context.Context, projectID uuid.UUID) ([]*project.ProjectFieldDevice, error) {
	var records []*ProjectFieldDeviceRecord
	err := r.db.WithContext(ctx).Where("project_id = ?", projectID).Find(&records).Error
	return toProjectFieldDeviceDomains(records), err
}

func (r *projectFieldDeviceRepo) GetByFieldDeviceID(ctx context.Context, fieldDeviceID uuid.UUID) ([]*project.ProjectFieldDevice, error) {
	return r.GetByFieldDeviceIDs(ctx, []uuid.UUID{fieldDeviceID})
}

func (r *projectFieldDeviceRepo) GetByFieldDeviceIDs(ctx context.Context, fieldDeviceIDs []uuid.UUID) ([]*project.ProjectFieldDevice, error) {
	if len(fieldDeviceIDs) == 0 {
		return []*project.ProjectFieldDevice{}, nil
	}

	var records []*ProjectFieldDeviceRecord
	err := r.db.WithContext(ctx).Where("field_device_id IN ?", fieldDeviceIDs).Find(&records).Error
	return toProjectFieldDeviceDomains(records), err
}

func (r *projectFieldDeviceRepo) DeleteByFieldDeviceIDs(ctx context.Context, fieldDeviceIDs []uuid.UUID) error {
	if len(fieldDeviceIDs) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).
		Where("field_device_id IN ?", fieldDeviceIDs).
		Delete(&ProjectFieldDeviceRecord{}).Error
}

func (r *projectFieldDeviceRepo) DeleteBySPSControllerSystemTypeIDs(ctx context.Context, systemTypeIDs []uuid.UUID) error {
	if len(systemTypeIDs) == 0 {
		return nil
	}

	const statement = `
		DELETE FROM project_field_devices
		USING field_devices
		WHERE project_field_devices.field_device_id = field_devices.id
			AND field_devices.sps_controller_system_type_id IN ?
	`

	for _, chunk := range uuidChunks(systemTypeIDs, projectLinkIDFilterChunkSize) {
		if err := r.db.WithContext(ctx).Exec(statement, chunk).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *projectFieldDeviceRepo) DeleteByProjectAndFieldDevice(ctx context.Context, projectID, fieldDeviceID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("project_id = ? AND field_device_id = ?", projectID, fieldDeviceID).
		Delete(&ProjectFieldDeviceRecord{}).Error
}
