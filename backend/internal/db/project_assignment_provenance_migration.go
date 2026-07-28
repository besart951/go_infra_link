package db

import (
	"fmt"

	"gorm.io/gorm"
)

const projectAssignmentSourceIDExpression = `(
	substr(source_hash.value, 1, 8) || '-' ||
	substr(source_hash.value, 9, 4) || '-' ||
	substr(source_hash.value, 13, 4) || '-' ||
	substr(source_hash.value, 17, 4) || '-' ||
	substr(source_hash.value, 21, 12)
)::uuid`

// migrateProjectAssignmentProvenance is the additive phase of the
// explicit-versus-inherited project-link rollout. Historical link rows cannot
// be classified reliably, so they are conservatively backfilled as explicit.
// That preserves access and makes any later cleanup fail safe.
func migrateProjectAssignmentProvenance(db *gorm.DB) error {
	if db == nil || db.Dialector == nil || db.Dialector.Name() != "postgres" {
		return nil
	}

	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS project_sps_controller_assignment_sources (
			id uuid PRIMARY KEY,
			created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
			project_sps_controller_id uuid NOT NULL,
			source_kind varchar(50) NOT NULL,
			source_entity_id uuid NOT NULL,
			CONSTRAINT fk_project_sps_assignment_source_link
				FOREIGN KEY (project_sps_controller_id)
				REFERENCES project_sps_controllers(id)
				ON UPDATE CASCADE
				ON DELETE CASCADE,
			CONSTRAINT chk_project_sps_assignment_source_kind
				CHECK (source_kind IN (
					'explicit',
					'control_cabinet',
					'sps_controller',
					'sps_controller_system_type'
				)),
			CONSTRAINT uq_project_sps_assignment_source
				UNIQUE (
					project_sps_controller_id,
					source_kind,
					source_entity_id
				)
		)
	`).Error; err != nil {
		return fmt.Errorf("create SPS project-assignment sources: %w", err)
	}

	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS project_field_device_assignment_sources (
			id uuid PRIMARY KEY,
			created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
			project_field_device_id uuid NOT NULL,
			source_kind varchar(50) NOT NULL,
			source_entity_id uuid NOT NULL,
			CONSTRAINT fk_project_field_assignment_source_link
				FOREIGN KEY (project_field_device_id)
				REFERENCES project_field_devices(id)
				ON UPDATE CASCADE
				ON DELETE CASCADE,
			CONSTRAINT chk_project_field_assignment_source_kind
				CHECK (source_kind IN (
					'explicit',
					'control_cabinet',
					'sps_controller',
					'sps_controller_system_type'
				)),
			CONSTRAINT uq_project_field_assignment_source
				UNIQUE (
					project_field_device_id,
					source_kind,
					source_entity_id
				)
		)
	`).Error; err != nil {
		return fmt.Errorf("create FieldDevice project-assignment sources: %w", err)
	}

	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_project_sps_assignment_source_lookup
		ON project_sps_controller_assignment_sources (
			source_kind,
			source_entity_id,
			project_sps_controller_id
		)
	`).Error; err != nil {
		return fmt.Errorf("index SPS project-assignment sources: %w", err)
	}
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_project_field_assignment_source_lookup
		ON project_field_device_assignment_sources (
			source_kind,
			source_entity_id,
			project_field_device_id
		)
	`).Error; err != nil {
		return fmt.Errorf("index FieldDevice project-assignment sources: %w", err)
	}

	if err := db.Exec(fmt.Sprintf(`
		INSERT INTO project_sps_controller_assignment_sources (
			id,
			project_sps_controller_id,
			source_kind,
			source_entity_id
		)
		SELECT
			%s,
			link.id,
			'explicit',
			link.sps_controller_id
		FROM project_sps_controllers AS link
		CROSS JOIN LATERAL (
			SELECT md5(
				'project-sps-assignment-source:' ||
				link.id::text ||
				':explicit:' ||
				link.sps_controller_id::text
			) AS value
		) AS source_hash
		ON CONFLICT (
			project_sps_controller_id,
			source_kind,
			source_entity_id
		) DO NOTHING
	`, projectAssignmentSourceIDExpression)).Error; err != nil {
		return fmt.Errorf("backfill SPS project-assignment sources: %w", err)
	}

	if err := db.Exec(fmt.Sprintf(`
		INSERT INTO project_field_device_assignment_sources (
			id,
			project_field_device_id,
			source_kind,
			source_entity_id
		)
		SELECT
			%s,
			link.id,
			'explicit',
			link.field_device_id
		FROM project_field_devices AS link
		CROSS JOIN LATERAL (
			SELECT md5(
				'project-field-assignment-source:' ||
				link.id::text ||
				':explicit:' ||
				link.field_device_id::text
			) AS value
		) AS source_hash
		ON CONFLICT (
			project_field_device_id,
			source_kind,
			source_entity_id
		) DO NOTHING
	`, projectAssignmentSourceIDExpression)).Error; err != nil {
		return fmt.Errorf("backfill FieldDevice project-assignment sources: %w", err)
	}

	return nil
}
