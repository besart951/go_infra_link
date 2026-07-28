package db

import (
	"fmt"

	"gorm.io/gorm"
)

const (
	fieldDevicePlacementConstraint = "uq_field_devices_placement_apparat_nr"
	spsDeviceNameNormalizedIndex   = "uq_sps_controllers_cabinet_device_name_normalized"
	spsGADeviceNormalizedIndex     = "uq_sps_controllers_cabinet_ga_device_normalized"
)

type fieldDevicePlacementDuplicate struct {
	SPSControllerSystemTypeID string `gorm:"column:sps_controller_system_type_id"`
	SystemPartID              string `gorm:"column:system_part_id"`
	ApparatID                 string `gorm:"column:apparat_id"`
	ApparatNr                 int    `gorm:"column:apparat_nr"`
	IDs                       string `gorm:"column:ids"`
}

// migrateFieldDevicePlacementUniqueness reports legacy conflicts without
// changing them, then installs a deferred constraint. Bulk updates execute in
// one application transaction, so a number permutation may temporarily
// collide while the final committed placement remains unique.
func migrateFieldDevicePlacementUniqueness(db *gorm.DB) error {
	if db == nil || db.Dialector == nil || db.Dialector.Name() != "postgres" {
		return nil
	}

	var invalidCount int64
	if err := db.Raw(`
		SELECT COUNT(*)
		FROM field_devices
		WHERE apparat_nr IS NULL OR apparat_nr < 1 OR apparat_nr > 99
	`).Scan(&invalidCount).Error; err != nil {
		return fmt.Errorf("inspect FieldDevice ApparatNr range: %w", err)
	}
	if invalidCount != 0 {
		return fmt.Errorf(
			"cannot install %s: field_devices contains %d null or out-of-range apparat_nr values",
			fieldDevicePlacementConstraint,
			invalidCount,
		)
	}

	var duplicate fieldDevicePlacementDuplicate
	result := db.Raw(`
		SELECT
			sps_controller_system_type_id::text AS sps_controller_system_type_id,
			system_part_id::text AS system_part_id,
			apparat_id::text AS apparat_id,
			apparat_nr,
			array_agg(id ORDER BY id)::text AS ids
		FROM field_devices
		GROUP BY
			sps_controller_system_type_id,
			system_part_id,
			apparat_id,
			apparat_nr
		HAVING COUNT(*) > 1
		ORDER BY
			sps_controller_system_type_id,
			system_part_id,
			apparat_id,
			apparat_nr
		LIMIT 1
	`).Scan(&duplicate)
	if result.Error != nil {
		return fmt.Errorf("inspect duplicate FieldDevice placements: %w", result.Error)
	}
	if result.RowsAffected != 0 {
		return fmt.Errorf(
			"cannot install %s: duplicate placement sps_controller_system_type_id=%s system_part_id=%s apparat_id=%s apparat_nr=%d ids=%s",
			fieldDevicePlacementConstraint,
			duplicate.SPSControllerSystemTypeID,
			duplicate.SystemPartID,
			duplicate.ApparatID,
			duplicate.ApparatNr,
			duplicate.IDs,
		)
	}

	var exists bool
	if err := db.Raw(`
		SELECT EXISTS (
			SELECT 1
			FROM pg_constraint
			WHERE conname = ?
			  AND conrelid = 'field_devices'::regclass
		)
	`, fieldDevicePlacementConstraint).Scan(&exists).Error; err != nil {
		return fmt.Errorf("inspect constraint %s: %w", fieldDevicePlacementConstraint, err)
	}
	if exists {
		return nil
	}

	return db.Exec(fmt.Sprintf(`
		ALTER TABLE field_devices
		ADD CONSTRAINT %s
		UNIQUE (
			sps_controller_system_type_id,
			system_part_id,
			apparat_id,
			apparat_nr
		)
		DEFERRABLE INITIALLY DEFERRED
	`, fieldDevicePlacementConstraint)).Error
}

type normalizedSPSDuplicate struct {
	ControlCabinetID string `gorm:"column:control_cabinet_id"`
	NormalizedValue  string `gorm:"column:normalized_value"`
	IDs              string `gorm:"column:ids"`
}

// migrateSPSControllerNormalizedUniqueness makes the database normalization
// match service comparisons. The audit fails with the conflicting IDs; no
// existing SPSController is renamed or deleted.
func migrateSPSControllerNormalizedUniqueness(db *gorm.DB) error {
	if db == nil || db.Dialector == nil || db.Dialector.Name() != "postgres" {
		return nil
	}

	if err := reportNormalizedSPSDuplicates(
		db,
		spsDeviceNameNormalizedIndex,
		"LOWER(BTRIM(device_name))",
		"TRUE",
	); err != nil {
		return err
	}
	if err := reportNormalizedSPSDuplicates(
		db,
		spsGADeviceNormalizedIndex,
		"UPPER(BTRIM(ga_device))",
		"ga_device IS NOT NULL AND BTRIM(ga_device) <> ''",
	); err != nil {
		return err
	}

	if err := db.Exec(fmt.Sprintf(`
		CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS %s
		ON sps_controllers (control_cabinet_id, LOWER(BTRIM(device_name)))
	`, spsDeviceNameNormalizedIndex)).Error; err != nil {
		return fmt.Errorf("create %s: %w", spsDeviceNameNormalizedIndex, err)
	}
	if err := db.Exec(fmt.Sprintf(`
		CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS %s
		ON sps_controllers (control_cabinet_id, UPPER(BTRIM(ga_device)))
		WHERE ga_device IS NOT NULL AND BTRIM(ga_device) <> ''
	`, spsGADeviceNormalizedIndex)).Error; err != nil {
		return fmt.Errorf("create %s: %w", spsGADeviceNormalizedIndex, err)
	}
	return nil
}

func reportNormalizedSPSDuplicates(
	db *gorm.DB,
	indexName string,
	normalizationExpression string,
	predicate string,
) error {
	var duplicate normalizedSPSDuplicate
	query := fmt.Sprintf(`
		SELECT
			control_cabinet_id::text AS control_cabinet_id,
			%s AS normalized_value,
			array_agg(id ORDER BY id)::text AS ids
		FROM sps_controllers
		WHERE %s
		GROUP BY control_cabinet_id, %s
		HAVING COUNT(*) > 1
		ORDER BY control_cabinet_id, %s
		LIMIT 1
	`, normalizationExpression, predicate, normalizationExpression, normalizationExpression)
	result := db.Raw(query).Scan(&duplicate)
	if result.Error != nil {
		return fmt.Errorf("inspect conflicts for %s: %w", indexName, result.Error)
	}
	if result.RowsAffected == 0 {
		return nil
	}
	return fmt.Errorf(
		"cannot install %s: normalized SPS value cabinet_id=%s value=%q ids=%s",
		indexName,
		duplicate.ControlCabinetID,
		duplicate.NormalizedValue,
		duplicate.IDs,
	)
}
