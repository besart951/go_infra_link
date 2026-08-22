package db

import (
	"fmt"
	"strings"

	facilitysql "github.com/besart951/go_infra_link/backend/internal/repository/facilitysql"
	"gorm.io/gorm"
)

type ownershipCheck struct {
	name      string
	statement string
}

var ownershipChecks = []ownershipCheck{
	{
		name: "BACnet objects with multiple ObjectData owners",
		statement: `SELECT bacnet_object_id FROM object_data_bacnet_objects
			GROUP BY bacnet_object_id HAVING COUNT(DISTINCT object_data_id) > 1 LIMIT 20`,
	},
	{
		name: "BACnet objects used as template and FieldDevice instance",
		statement: `SELECT DISTINCT bo.id FROM bacnet_objects bo
			JOIN object_data_bacnet_objects odb ON odb.bacnet_object_id = bo.id
			WHERE bo.field_device_id IS NOT NULL LIMIT 20`,
	},
	{
		name: "template software references outside their ObjectData owner",
		statement: `SELECT DISTINCT source.bacnet_object_id FROM object_data_bacnet_objects source
			JOIN bacnet_objects bo ON bo.id = source.bacnet_object_id
			LEFT JOIN object_data_bacnet_objects target
			  ON target.bacnet_object_id = bo.software_reference_id
			 AND target.object_data_id = source.object_data_id
			WHERE bo.software_reference_id IS NOT NULL AND target.bacnet_object_id IS NULL LIMIT 20`,
	},
	{
		name: "contradicting FieldDevice specification ownership",
		statement: `SELECT DISTINCT fd.id FROM field_devices fd
			JOIN specifications specification ON specification.id = fd.specification_id
			WHERE specification.field_device_id IS NOT NULL
			  AND specification.field_device_id <> fd.id LIMIT 20`,
	},
}

func migrateFacilityOwnership(db *gorm.DB) error {
	legacy, err := hasLegacyFacilityOwnership(db)
	if err != nil {
		return err
	}
	if legacy {
		if err := preflightFacilityOwnership(db); err != nil {
			return err
		}
	}
	if err := db.AutoMigrate(
		&facilitysql.BacnetObjectTemplateRecord{},
		&facilitysql.BacnetObjectTemplateAlarmValueRecord{},
	); err != nil {
		return err
	}
	if legacy {
		if err := backfillCanonicalSpecificationOwner(db); err != nil {
			return err
		}
		if err := backfillBacnetTemplates(db); err != nil {
			return err
		}
	}
	return addFacilityOwnershipConstraints(db)
}

func hasLegacyFacilityOwnership(db *gorm.DB) (bool, error) {
	hasJoinTable := db.Migrator().HasTable("object_data_bacnet_objects")
	hasSpecificationColumn := db.Migrator().HasColumn("field_devices", "specification_id")
	if hasJoinTable != hasSpecificationColumn {
		return false, fmt.Errorf("incomplete legacy Facility ownership schema")
	}
	return hasJoinTable, nil
}

func preflightFacilityOwnership(db *gorm.DB) error {
	for _, check := range ownershipChecks {
		ids, err := scanOwnershipConflictIDs(db, check.statement)
		if err != nil {
			return fmt.Errorf("facility ownership preflight %q: %w", check.name, err)
		}
		if len(ids) > 0 {
			return fmt.Errorf("facility ownership preflight failed: %s: %s", check.name, strings.Join(ids, ", "))
		}
	}
	return nil
}

func scanOwnershipConflictIDs(db *gorm.DB, statement string) ([]string, error) {
	rows, err := db.Raw(statement).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ids := make([]string, 0, 20)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func backfillCanonicalSpecificationOwner(db *gorm.DB) error {
	return db.Exec(`UPDATE specifications
		SET field_device_id = (
			SELECT fd.id FROM field_devices fd WHERE fd.specification_id = specifications.id
		)
		WHERE field_device_id IS NULL
		  AND EXISTS (SELECT 1 FROM field_devices fd WHERE fd.specification_id = specifications.id)`).Error
}

func backfillBacnetTemplates(db *gorm.DB) error {
	statement := `INSERT INTO bacnet_object_templates (
		id, created_at, updated_at, version, object_data_id, text_fix, description,
		gms_visible, optional, text_individual, software_type, software_number,
		hardware_type, hardware_quantity, software_reference_id, state_text_id,
		notification_class_id, alarm_type_id
	)
	SELECT bo.id, bo.created_at, bo.updated_at, bo.version, odb.object_data_id,
		bo.text_fix, bo.description, bo.gms_visible, bo.optional, bo.text_individual,
		bo.software_type, bo.software_number, bo.hardware_type, bo.hardware_quantity,
		bo.software_reference_id, bo.state_text_id, bo.notification_class_id, bo.alarm_type_id
	FROM bacnet_objects bo
	JOIN object_data_bacnet_objects odb ON odb.bacnet_object_id = bo.id
	WHERE NOT EXISTS (SELECT 1 FROM bacnet_object_templates target WHERE target.id = bo.id)`
	if err := db.Exec(statement).Error; err != nil {
		return err
	}
	return backfillBacnetTemplateAlarmValues(db)
}

func backfillBacnetTemplateAlarmValues(db *gorm.DB) error {
	return db.Exec(`INSERT INTO bacnet_object_template_alarm_values (
		id, created_at, updated_at, version, template_id, alarm_type_field_id,
		value_number, value_integer, value_boolean, value_string, value_json, unit_id, source
	)
	SELECT alarm_value.id, alarm_value.created_at, alarm_value.updated_at, alarm_value.version,
		alarm_value.bacnet_object_id, alarm_value.alarm_type_field_id, alarm_value.value_number,
		alarm_value.value_integer, alarm_value.value_boolean, alarm_value.value_string, alarm_value.value_json,
		alarm_value.unit_id, alarm_value.source
	FROM bacnet_object_alarm_values alarm_value
	JOIN bacnet_object_templates template ON template.id = alarm_value.bacnet_object_id
	WHERE NOT EXISTS (
		SELECT 1 FROM bacnet_object_template_alarm_values target WHERE target.id = alarm_value.id
	)`).Error
}

func addFacilityOwnershipConstraints(db *gorm.DB) error {
	if db.Dialector == nil || db.Dialector.Name() != "postgres" {
		return nil
	}
	statements := []string{
		`DO $$ BEGIN IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_bacnet_template_owner') THEN
			ALTER TABLE bacnet_object_templates ADD CONSTRAINT fk_bacnet_template_owner
			FOREIGN KEY (object_data_id) REFERENCES object_data(id) ON DELETE CASCADE; END IF; END $$`,
		`DO $$ BEGIN IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_bacnet_template_software_reference') THEN
			ALTER TABLE bacnet_object_templates ADD CONSTRAINT fk_bacnet_template_software_reference
			FOREIGN KEY (object_data_id, software_reference_id)
			REFERENCES bacnet_object_templates(object_data_id, id) DEFERRABLE INITIALLY DEFERRED; END IF; END $$`,
		`DO $$ BEGIN IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_bacnet_template_alarm_owner') THEN
			ALTER TABLE bacnet_object_template_alarm_values ADD CONSTRAINT fk_bacnet_template_alarm_owner
			FOREIGN KEY (template_id) REFERENCES bacnet_object_templates(id) ON DELETE CASCADE; END IF; END $$`,
		`DO $$ BEGIN IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'uq_bacnet_objects_owner_id') THEN
			ALTER TABLE bacnet_objects ADD CONSTRAINT uq_bacnet_objects_owner_id
			UNIQUE (field_device_id,id); END IF; END $$`,
		`DO $$ BEGIN IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_bacnet_objects_owner_reference') THEN
			ALTER TABLE bacnet_objects ADD CONSTRAINT fk_bacnet_objects_owner_reference
			FOREIGN KEY (field_device_id,software_reference_id)
			REFERENCES bacnet_objects(field_device_id,id) DEFERRABLE INITIALLY DEFERRED; END IF; END $$`,
	}
	for _, statement := range statements {
		if err := db.Exec(statement).Error; err != nil {
			return err
		}
	}
	return nil
}
