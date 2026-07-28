package projectsql

import (
	"context"
	"fmt"
	"sort"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const projectAssignmentSourceBatchSize = 100

const deterministicAssignmentSourceIDExpression = `(
	substr(source_hash.value, 1, 8) || '-' ||
	substr(source_hash.value, 9, 4) || '-' ||
	substr(source_hash.value, 13, 4) || '-' ||
	substr(source_hash.value, 17, 4) || '-' ||
	substr(source_hash.value, 21, 12)
)::uuid`

func supportsProjectAssignmentProvenance(db *gorm.DB) bool {
	return db != nil && db.Dialector != nil && db.Dialector.Name() == "postgres"
}

func (r *projectSPSControllerRepo) AssignmentProvenanceEnabled() bool {
	return supportsProjectAssignmentProvenance(r.db)
}

func (r *projectFieldDeviceRepo) AssignmentProvenanceEnabled() bool {
	return supportsProjectAssignmentProvenance(r.db)
}

func (r *projectSPSControllerRepo) AddAssignmentSource(
	ctx context.Context,
	projectID uuid.UUID,
	spsControllerIDs []uuid.UUID,
	source domainProject.AssignmentSource,
) error {
	if len(spsControllerIDs) == 0 || !supportsProjectAssignmentProvenance(r.db) {
		return nil
	}
	if err := source.Validate(); err != nil {
		return err
	}

	const statement = `
		INSERT INTO project_sps_controller_assignment_sources (
			id,
			project_sps_controller_id,
			source_kind,
			source_entity_id
		)
		SELECT
			` + deterministicAssignmentSourceIDExpression + `,
			link.id,
			?,
			source_value.entity_id
		FROM project_sps_controllers AS link
		CROSS JOIN LATERAL (
			SELECT CASE
				WHEN ? = 'explicit' THEN link.sps_controller_id
				ELSE ?::uuid
			END AS entity_id
		) AS source_value
		CROSS JOIN LATERAL (
			SELECT md5(
				'project-sps-assignment-source:' ||
				link.id::text ||
				':' || ? ||
				':' || source_value.entity_id::text
			) AS value
		) AS source_hash
		WHERE link.project_id = ?
		  AND link.sps_controller_id IN ?
		ON CONFLICT (
			project_sps_controller_id,
			source_kind,
			source_entity_id
		) DO NOTHING
	`

	for _, chunk := range uuidChunks(spsControllerIDs, projectLinkIDFilterChunkSize) {
		if err := r.db.WithContext(ctx).Exec(
			statement,
			source.Kind,
			source.Kind,
			source.SourceEntityID,
			source.Kind,
			projectID,
			chunk,
		).Error; err != nil {
			return fmt.Errorf("persist SPS project-assignment source: %w", err)
		}
	}
	return nil
}

func (r *projectSPSControllerRepo) ListProjectIDsByAssignmentSource(
	ctx context.Context,
	source domainProject.AssignmentSource,
) ([]uuid.UUID, error) {
	if !supportsProjectAssignmentProvenance(r.db) {
		return nil, nil
	}
	if err := validateAssignmentSourceRemoval(source); err != nil {
		return nil, err
	}

	var projectIDs []uuid.UUID
	if err := r.db.WithContext(ctx).
		Table("project_sps_controllers AS link").
		Distinct("link.project_id").
		Joins(`
			JOIN project_sps_controller_assignment_sources AS source
			  ON source.project_sps_controller_id = link.id
		`).
		Where(
			"source.source_kind = ? AND source.source_entity_id = ?",
			source.Kind,
			source.SourceEntityID,
		).
		Order("link.project_id ASC").
		Pluck("link.project_id", &projectIDs).Error; err != nil {
		return nil, fmt.Errorf("list SPS project-assignment source projects: %w", err)
	}
	return projectIDs, nil
}

func (r *projectFieldDeviceRepo) AddAssignmentSource(
	ctx context.Context,
	projectID uuid.UUID,
	fieldDeviceIDs []uuid.UUID,
	source domainProject.AssignmentSource,
) error {
	if len(fieldDeviceIDs) == 0 || !supportsProjectAssignmentProvenance(r.db) {
		return nil
	}
	if err := source.Validate(); err != nil {
		return err
	}

	const statement = `
		INSERT INTO project_field_device_assignment_sources (
			id,
			project_field_device_id,
			source_kind,
			source_entity_id
		)
		SELECT
			` + deterministicAssignmentSourceIDExpression + `,
			link.id,
			?,
			source_value.entity_id
		FROM project_field_devices AS link
		CROSS JOIN LATERAL (
			SELECT CASE
				WHEN ? = 'explicit' THEN link.field_device_id
				ELSE ?::uuid
			END AS entity_id
		) AS source_value
		CROSS JOIN LATERAL (
			SELECT md5(
				'project-field-assignment-source:' ||
				link.id::text ||
				':' || ? ||
				':' || source_value.entity_id::text
			) AS value
		) AS source_hash
		WHERE link.project_id = ?
		  AND link.field_device_id IN ?
		ON CONFLICT (
			project_field_device_id,
			source_kind,
			source_entity_id
		) DO NOTHING
	`

	for _, chunk := range uuidChunks(fieldDeviceIDs, projectLinkIDFilterChunkSize) {
		if err := r.db.WithContext(ctx).Exec(
			statement,
			source.Kind,
			source.Kind,
			source.SourceEntityID,
			source.Kind,
			projectID,
			chunk,
		).Error; err != nil {
			return fmt.Errorf("persist FieldDevice project-assignment source: %w", err)
		}
	}
	return nil
}

func (r *projectSPSControllerRepo) RemoveAssignmentSourceBatch(
	ctx context.Context,
	projectID uuid.UUID,
	source domainProject.AssignmentSource,
	afterID uuid.UUID,
	limit int,
) ([]uuid.UUID, []uuid.UUID, error) {
	if !supportsProjectAssignmentProvenance(r.db) {
		return nil, nil, nil
	}
	if err := validateAssignmentSourceRemoval(source); err != nil {
		return nil, nil, err
	}
	limit = normalizeAssignmentSourceBatchLimit(limit)

	const statement = `
		WITH targets AS (
			SELECT source.project_sps_controller_id AS link_id
			FROM project_sps_controller_assignment_sources AS source
			JOIN project_sps_controllers AS link
			  ON link.id = source.project_sps_controller_id
			WHERE link.project_id = ?
			  AND source.source_kind = ?
			  AND source.source_entity_id = ?
			  AND source.project_sps_controller_id > ?
			ORDER BY source.project_sps_controller_id ASC
			LIMIT ?
			FOR UPDATE OF source
		),
		deleted AS (
			DELETE FROM project_sps_controller_assignment_sources AS source
			USING targets
			WHERE source.project_sps_controller_id = targets.link_id
			  AND source.source_kind = ?
			  AND source.source_entity_id = ?
			RETURNING source.project_sps_controller_id AS link_id
		)
		SELECT DISTINCT link_id
		FROM deleted
		ORDER BY link_id ASC
	`
	processed, err := removeAssignmentSourceRows(
		ctx,
		r.db,
		statement,
		projectID,
		source,
		afterID,
		limit,
	)
	if err != nil || len(processed) == 0 {
		return processed, nil, err
	}
	unclaimed, err := listUnclaimedProjectLinks(
		ctx,
		r.db,
		"project_sps_controllers",
		"project_sps_controller_assignment_sources",
		"project_sps_controller_id",
		processed,
	)
	return processed, unclaimed, err
}

func (r *projectFieldDeviceRepo) RemoveAssignmentSourceBatch(
	ctx context.Context,
	projectID uuid.UUID,
	source domainProject.AssignmentSource,
	afterID uuid.UUID,
	limit int,
) ([]uuid.UUID, []uuid.UUID, error) {
	if !supportsProjectAssignmentProvenance(r.db) {
		return nil, nil, nil
	}
	if err := validateAssignmentSourceRemoval(source); err != nil {
		return nil, nil, err
	}
	limit = normalizeAssignmentSourceBatchLimit(limit)

	const statement = `
		WITH targets AS (
			SELECT source.project_field_device_id AS link_id
			FROM project_field_device_assignment_sources AS source
			JOIN project_field_devices AS link
			  ON link.id = source.project_field_device_id
			WHERE link.project_id = ?
			  AND source.source_kind = ?
			  AND source.source_entity_id = ?
			  AND source.project_field_device_id > ?
			ORDER BY source.project_field_device_id ASC
			LIMIT ?
			FOR UPDATE OF source
		),
		deleted AS (
			DELETE FROM project_field_device_assignment_sources AS source
			USING targets
			WHERE source.project_field_device_id = targets.link_id
			  AND source.source_kind = ?
			  AND source.source_entity_id = ?
			RETURNING source.project_field_device_id AS link_id
		)
		SELECT DISTINCT link_id
		FROM deleted
		ORDER BY link_id ASC
	`
	processed, err := removeAssignmentSourceRows(
		ctx,
		r.db,
		statement,
		projectID,
		source,
		afterID,
		limit,
	)
	if err != nil || len(processed) == 0 {
		return processed, nil, err
	}
	unclaimed, err := listUnclaimedProjectLinks(
		ctx,
		r.db,
		"project_field_devices",
		"project_field_device_assignment_sources",
		"project_field_device_id",
		processed,
	)
	return processed, unclaimed, err
}

func (r *projectSPSControllerRepo) HasAssignmentSourceOtherThan(
	ctx context.Context,
	linkID uuid.UUID,
	allowed domainProject.AssignmentSource,
) (bool, error) {
	return hasAssignmentSourceOtherThan(
		ctx,
		r.db,
		"project_sps_controller_assignment_sources",
		"project_sps_controller_id",
		linkID,
		allowed,
	)
}

func (r *projectFieldDeviceRepo) HasAssignmentSourceOtherThan(
	ctx context.Context,
	linkID uuid.UUID,
	allowed domainProject.AssignmentSource,
) (bool, error) {
	return hasAssignmentSourceOtherThan(
		ctx,
		r.db,
		"project_field_device_assignment_sources",
		"project_field_device_id",
		linkID,
		allowed,
	)
}

func (r *projectSPSControllerRepo) ReplaceExplicitAssignmentSource(
	ctx context.Context,
	linkID uuid.UUID,
	fromEntityID uuid.UUID,
	toEntityID uuid.UUID,
) error {
	return replaceExplicitAssignmentSource(
		ctx,
		r.db,
		"project_sps_controller_assignment_sources",
		"project_sps_controller_id",
		linkID,
		fromEntityID,
		toEntityID,
	)
}

func (r *projectFieldDeviceRepo) ReplaceExplicitAssignmentSource(
	ctx context.Context,
	linkID uuid.UUID,
	fromEntityID uuid.UUID,
	toEntityID uuid.UUID,
) error {
	return replaceExplicitAssignmentSource(
		ctx,
		r.db,
		"project_field_device_assignment_sources",
		"project_field_device_id",
		linkID,
		fromEntityID,
		toEntityID,
	)
}

func validateAssignmentSourceRemoval(source domainProject.AssignmentSource) error {
	if err := source.Validate(); err != nil {
		return err
	}
	if source.SourceEntityID == uuid.Nil {
		return fmt.Errorf(
			"assignment source removal requires source entity: %w",
			domain.ErrInvalidArgument,
		)
	}
	return nil
}

func normalizeAssignmentSourceBatchLimit(limit int) int {
	if limit <= 0 || limit > projectAssignmentSourceBatchSize {
		return projectAssignmentSourceBatchSize
	}
	return limit
}

func removeAssignmentSourceRows(
	ctx context.Context,
	db *gorm.DB,
	statement string,
	projectID uuid.UUID,
	source domainProject.AssignmentSource,
	afterID uuid.UUID,
	limit int,
) ([]uuid.UUID, error) {
	var rows []struct {
		LinkID uuid.UUID `gorm:"column:link_id"`
	}
	if err := db.WithContext(ctx).Raw(
		statement,
		projectID,
		source.Kind,
		source.SourceEntityID,
		afterID,
		limit,
		source.Kind,
		source.SourceEntityID,
	).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("remove project-assignment source: %w", err)
	}
	ids := make([]uuid.UUID, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.LinkID)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i].String() < ids[j].String() })
	return ids, nil
}

func listUnclaimedProjectLinks(
	ctx context.Context,
	db *gorm.DB,
	linkTable string,
	sourceTable string,
	linkColumn string,
	linkIDs []uuid.UUID,
) ([]uuid.UUID, error) {
	if len(linkIDs) == 0 {
		return nil, nil
	}
	statement := fmt.Sprintf(`
		SELECT link.id
		FROM %s AS link
		WHERE link.id IN ?
		  AND NOT EXISTS (
			SELECT 1
			FROM %s AS source
			WHERE source.%s = link.id
		)
		ORDER BY link.id ASC
	`, linkTable, sourceTable, linkColumn)
	var ids []uuid.UUID
	if err := db.WithContext(ctx).Raw(statement, linkIDs).Scan(&ids).Error; err != nil {
		return nil, fmt.Errorf("list unclaimed project links: %w", err)
	}
	return ids, nil
}

func hasAssignmentSourceOtherThan(
	ctx context.Context,
	db *gorm.DB,
	sourceTable string,
	linkColumn string,
	linkID uuid.UUID,
	allowed domainProject.AssignmentSource,
) (bool, error) {
	if !supportsProjectAssignmentProvenance(db) {
		return false, nil
	}
	if err := validateAssignmentSourceRemoval(allowed); err != nil {
		return false, err
	}
	statement := fmt.Sprintf(`
		SELECT EXISTS (
			SELECT 1
			FROM %s
			WHERE %s = ?
			  AND (
				source_kind <> ? OR
				source_entity_id <> ?
			  )
		)
	`, sourceTable, linkColumn)
	var exists bool
	if err := db.WithContext(ctx).Raw(
		statement,
		linkID,
		allowed.Kind,
		allowed.SourceEntityID,
	).Scan(&exists).Error; err != nil {
		return false, fmt.Errorf("inspect project-assignment sources: %w", err)
	}
	return exists, nil
}

func replaceExplicitAssignmentSource(
	ctx context.Context,
	db *gorm.DB,
	sourceTable string,
	linkColumn string,
	linkID uuid.UUID,
	fromEntityID uuid.UUID,
	toEntityID uuid.UUID,
) error {
	if !supportsProjectAssignmentProvenance(db) {
		return nil
	}
	if linkID == uuid.Nil || fromEntityID == uuid.Nil || toEntityID == uuid.Nil {
		return domain.ErrInvalidArgument
	}
	statement := fmt.Sprintf(`
		UPDATE %s
		SET source_entity_id = ?
		WHERE %s = ?
		  AND source_kind = 'explicit'
		  AND source_entity_id = ?
	`, sourceTable, linkColumn)
	result := db.WithContext(ctx).Exec(
		statement,
		toEntityID,
		linkID,
		fromEntityID,
	)
	if result.Error != nil {
		return fmt.Errorf("replace explicit project-assignment source: %w", result.Error)
	}
	if result.RowsAffected != 1 {
		return fmt.Errorf(
			"replace explicit project-assignment source for link %s: %w",
			linkID,
			domain.ErrConflict,
		)
	}
	return nil
}
