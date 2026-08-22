package db

import (
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

const facilityContractVersion = "202609050001"

type FacilityContractOptions struct {
	CompatibleReleaseDelivered bool
	ApplicationsStopped        bool
	LegacyIdleSince            time.Time
	BackupVerified             bool
}

type contractCheck struct {
	name       string
	sql        string
	legacyOnly bool
}

var facilityContractChecks = []contractCheck{
	{name: "specifications without owner", sql: `SELECT id FROM specifications WHERE field_device_id IS NULL LIMIT 20`},
	{name: "duplicate specification owners", sql: `SELECT field_device_id FROM specifications GROUP BY field_device_id HAVING COUNT(*) > 1 LIMIT 20`},
	{name: "contradicting specification owners", sql: `SELECT fd.id FROM field_devices fd JOIN specifications s ON s.id=fd.specification_id WHERE s.field_device_id<>fd.id LIMIT 20`, legacyOnly: true},
	{name: "mixed BACnet template and instance use", sql: `SELECT bo.id FROM bacnet_objects bo JOIN object_data_bacnet_objects odb ON odb.bacnet_object_id=bo.id WHERE bo.field_device_id IS NOT NULL LIMIT 20`, legacyOnly: true},
	{name: "BACnet templates with multiple owners", sql: `SELECT bacnet_object_id FROM object_data_bacnet_objects GROUP BY bacnet_object_id HAVING COUNT(DISTINCT object_data_id)<>1 LIMIT 20`, legacyOnly: true},
	{name: "missing canonical BACnet templates", sql: `SELECT odb.bacnet_object_id FROM object_data_bacnet_objects odb LEFT JOIN bacnet_object_templates t ON t.id=odb.bacnet_object_id AND t.object_data_id=odb.object_data_id WHERE t.id IS NULL LIMIT 20`, legacyOnly: true},
	{name: "different canonical BACnet template attributes", sql: templateAttributeConflictSQL, legacyOnly: true},
	{name: "different canonical BACnet alarm values", sql: templateAlarmConflictSQL, legacyOnly: true},
	{name: "template references outside owner", sql: `SELECT source.id FROM bacnet_object_templates source JOIN bacnet_object_templates target ON target.id=source.software_reference_id WHERE source.software_reference_id IS NOT NULL AND source.object_data_id<>target.object_data_id LIMIT 20`},
	{name: "instance references outside owner", sql: `SELECT source.id FROM bacnet_objects source JOIN bacnet_objects target ON target.id=source.software_reference_id WHERE source.software_reference_id IS NOT NULL AND source.field_device_id IS DISTINCT FROM target.field_device_id LIMIT 20`},
	{name: "BACnet instances without owner", sql: `SELECT id FROM bacnet_objects WHERE field_device_id IS NULL AND id NOT IN (SELECT bacnet_object_id FROM object_data_bacnet_objects) LIMIT 20`, legacyOnly: true},
	{name: "FieldDevices without number owner", sql: `SELECT id FROM field_devices WHERE sps_controller_system_type_id IS NULL OR system_part_id IS NULL OR apparat_id IS NULL OR apparat_nr IS NULL LIMIT 20`},
	{name: "duplicate FieldDevice number keys", sql: `SELECT min(id::text) FROM field_devices GROUP BY sps_controller_system_type_id,system_part_id,apparat_id,apparat_nr HAVING COUNT(*)>1 LIMIT 20`},
}

const templateAttributeConflictSQL = `SELECT bo.id FROM bacnet_objects bo
	JOIN object_data_bacnet_objects odb ON odb.bacnet_object_id=bo.id
	JOIN bacnet_object_templates t ON t.id=bo.id AND t.object_data_id=odb.object_data_id
	WHERE ROW(bo.text_fix,bo.description,bo.gms_visible,bo.optional,bo.text_individual,bo.software_type,
		bo.software_number,bo.hardware_type,bo.hardware_quantity,bo.software_reference_id,bo.state_text_id,
		bo.notification_class_id,bo.alarm_type_id) IS DISTINCT FROM
		ROW(t.text_fix,t.description,t.gms_visible,t.optional,t.text_individual,t.software_type,
		t.software_number,t.hardware_type,t.hardware_quantity,t.software_reference_id,t.state_text_id,
		t.notification_class_id,t.alarm_type_id) LIMIT 20`

const templateAlarmConflictSQL = `SELECT t.id FROM bacnet_object_templates t
	WHERE (SELECT md5(COALESCE(string_agg(jsonb_build_array(v.id,v.version,v.alarm_type_field_id,v.value_number,v.value_integer,v.value_boolean,v.value_string,v.value_json,v.unit_id,v.source)::text,'' ORDER BY v.id),'')) FROM bacnet_object_alarm_values v WHERE v.bacnet_object_id=t.id)
	IS DISTINCT FROM
	(SELECT md5(COALESCE(string_agg(jsonb_build_array(v.id,v.version,v.alarm_type_field_id,v.value_number,v.value_integer,v.value_boolean,v.value_string,v.value_json,v.unit_id,v.source)::text,'' ORDER BY v.id),'')) FROM bacnet_object_template_alarm_values v WHERE v.template_id=t.id) LIMIT 20`

func ApplyFacilityContractMigration(database *gorm.DB, options FacilityContractOptions) error {
	if err := validateContractOptions(database, options); err != nil {
		return err
	}
	applied, err := facilityContractApplied(database)
	if err != nil || applied {
		return err
	}
	return database.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(`SELECT pg_advisory_xact_lock(hashtext('facility-contract-v1'))`).Error; err != nil {
			return err
		}
		if err := ensureNoActiveFacilityJobs(tx); err != nil {
			return err
		}
		legacy, err := hasLegacyFacilityOwnership(tx)
		if err != nil {
			return err
		}
		if err := runFacilityContractPreflight(tx, legacy); err != nil {
			return err
		}
		if err := executeFacilityContractDDL(tx, legacy); err != nil {
			return err
		}
		if err := runFacilityContractPostflight(tx); err != nil {
			return err
		}
		return recordFacilityContract(tx)
	})
}

func facilityContractApplied(database *gorm.DB) (bool, error) {
	var count int64
	err := database.Model(&schemaMigration{}).Where("version = ?", facilityContractVersion).Count(&count).Error
	return count > 0, err
}

func validateContractOptions(database *gorm.DB, options FacilityContractOptions) error {
	if database == nil || database.Dialector == nil || database.Dialector.Name() != "postgres" {
		return fmt.Errorf("facility contract migration requires PostgreSQL")
	}
	if !options.CompatibleReleaseDelivered || !options.ApplicationsStopped || !options.BackupVerified {
		return fmt.Errorf("compatible release, stopped applications, and verified backup are required")
	}
	if options.LegacyIdleSince.IsZero() || time.Since(options.LegacyIdleSince) < 14*24*time.Hour {
		return fmt.Errorf("14 legacy-free days are required")
	}
	return nil
}

func ensureNoActiveFacilityJobs(tx *gorm.DB) error {
	var count int64
	if err := tx.Table("facility_jobs").Where("status IN ?", []string{"queued", "running"}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("facility contract blocked by %d active jobs", count)
	}
	return nil
}

func runFacilityContractPreflight(tx *gorm.DB, legacy bool) error {
	for _, check := range facilityContractChecks {
		if check.legacyOnly && !legacy {
			continue
		}
		ids, err := scanOwnershipConflictIDs(tx, check.sql)
		if err != nil {
			return fmt.Errorf("contract preflight %q: %w", check.name, err)
		}
		if len(ids) > 0 {
			return fmt.Errorf("contract preflight failed: %s: %s", check.name, strings.Join(ids, ", "))
		}
	}
	if !legacy {
		ids, err := scanOwnershipConflictIDs(tx, `SELECT id FROM bacnet_objects WHERE field_device_id IS NULL LIMIT 20`)
		if err != nil {
			return fmt.Errorf("contract preflight %q: %w", "BACnet instances without owner", err)
		}
		if len(ids) > 0 {
			return fmt.Errorf("contract preflight failed: BACnet instances without owner: %s", strings.Join(ids, ", "))
		}
	}
	return nil
}

func executeFacilityContractDDL(tx *gorm.DB, legacy bool) error {
	statements := make([]string, 0, 16)
	if legacy {
		statements = append(statements,
			`CREATE TEMP TABLE facility_legacy_template_ids ON COMMIT DROP AS SELECT bacnet_object_id AS id FROM object_data_bacnet_objects`,
			`DELETE FROM bacnet_object_alarm_values WHERE bacnet_object_id IN (SELECT id FROM facility_legacy_template_ids)`,
			`DELETE FROM bacnet_objects WHERE id IN (SELECT id FROM facility_legacy_template_ids)`,
			`DROP TABLE object_data_bacnet_objects`,
			`ALTER TABLE field_devices DROP CONSTRAINT IF EXISTS fk_field_devices_specification`,
			`DROP INDEX IF EXISTS idx_field_devices_specification_id`,
			`ALTER TABLE field_devices DROP COLUMN specification_id`,
		)
	}
	statements = append(statements,
		`ALTER TABLE specifications ALTER COLUMN field_device_id SET NOT NULL`,
		`ALTER TABLE bacnet_objects ALTER COLUMN field_device_id SET NOT NULL`,
		`ALTER TABLE specifications ADD CONSTRAINT uq_specifications_field_device UNIQUE (field_device_id)`,
		`ALTER TABLE bacnet_objects DROP CONSTRAINT IF EXISTS fk_bacnet_objects_software_reference`,
		`ALTER TABLE field_devices ADD CONSTRAINT uq_field_devices_number_scope UNIQUE (sps_controller_system_type_id,system_part_id,apparat_id,apparat_nr) DEFERRABLE INITIALLY DEFERRED`,
	)
	for _, statement := range statements {
		if err := tx.Exec(statement).Error; err != nil {
			return err
		}
	}
	return addFacilityOwnershipConstraints(tx)
}

func runFacilityContractPostflight(tx *gorm.DB) error {
	checks := []contractCheck{
		{name: "legacy join table remains", sql: `SELECT table_name FROM information_schema.tables WHERE table_schema=current_schema() AND table_name='object_data_bacnet_objects'`},
		{name: "legacy specification column remains", sql: `SELECT column_name FROM information_schema.columns WHERE table_schema=current_schema() AND table_name='field_devices' AND column_name='specification_id'`},
		{name: "required contract constraint missing", sql: `SELECT required.name FROM (VALUES
			('uq_specifications_field_device',false),
			('fk_bacnet_template_software_reference',true),
			('fk_bacnet_objects_owner_reference',true),
			('uq_field_devices_number_scope',true)
		) required(name,must_defer) LEFT JOIN pg_constraint c ON c.conname=required.name
		WHERE c.oid IS NULL OR (required.must_defer AND NOT c.condeferrable)`},
	}
	for _, check := range checks {
		ids, err := scanOwnershipConflictIDs(tx, check.sql)
		if err != nil {
			return err
		}
		if len(ids) > 0 {
			return fmt.Errorf("contract postflight failed: %s: %s", check.name, strings.Join(ids, ", "))
		}
	}
	return nil
}

func recordFacilityContract(tx *gorm.DB) error {
	entry := schemaMigration{Version: facilityContractVersion, Description: "facility_contract_remove_legacy_ownership", AppliedAt: time.Now().UTC()}
	return tx.Create(&entry).Error
}
