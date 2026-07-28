package db

import (
	"fmt"

	"gorm.io/gorm"
)

const (
	objectDataBacnetSoftwareIndex = "uq_object_data_bacnet_software"
	objectDataBacnetSyncFunction  = "sync_object_data_bacnet_software_key"
	bacnetObjectTemplateFunction  = "sync_bacnet_object_template_software_keys"
)

type objectDataBacnetSoftwareDuplicate struct {
	ObjectDataID           string `gorm:"column:object_data_id"`
	SoftwareTypeNormalized string `gorm:"column:software_type_normalized"`
	SoftwareNumber         int    `gorm:"column:software_number"`
	BacnetObjectIDs        string `gorm:"column:bacnet_object_ids"`
}

// migrateBacnetTemplateUniqueness denormalizes the template software key onto
// the join table. PostgreSQL can then enforce ObjectData-scoped uniqueness with
// a normal unique index under concurrent inserts and updates. TextFix is
// intentionally absent: duplicate TextFix values are a supported domain case.
func migrateBacnetTemplateUniqueness(db *gorm.DB) error {
	if db == nil || db.Dialector == nil || db.Dialector.Name() != "postgres" {
		return nil
	}

	if err := db.Exec(`
		ALTER TABLE object_data_bacnet_objects
			ADD COLUMN IF NOT EXISTS software_type_normalized varchar(50),
			ADD COLUMN IF NOT EXISTS software_number integer
	`).Error; err != nil {
		return fmt.Errorf("add ObjectData BACnet software-key columns: %w", err)
	}

	if err := db.Exec(fmt.Sprintf(`
		CREATE OR REPLACE FUNCTION %s()
		RETURNS trigger
		LANGUAGE plpgsql
		AS $$
		BEGIN
			SELECT LOWER(BTRIM(software_type)), software_number
			INTO NEW.software_type_normalized, NEW.software_number
			FROM bacnet_objects
			WHERE id = NEW.bacnet_object_id;

			IF NOT FOUND THEN
				RAISE EXCEPTION 'BACnet object %% does not exist', NEW.bacnet_object_id
					USING ERRCODE = '23503';
			END IF;
			RETURN NEW;
		END;
		$$
	`, objectDataBacnetSyncFunction)).Error; err != nil {
		return fmt.Errorf("create ObjectData BACnet software-key function: %w", err)
	}
	if err := db.Exec(fmt.Sprintf(`
		DROP TRIGGER IF EXISTS trg_object_data_bacnet_software_key
		ON object_data_bacnet_objects;
		CREATE TRIGGER trg_object_data_bacnet_software_key
		BEFORE INSERT OR UPDATE OF bacnet_object_id
		ON object_data_bacnet_objects
		FOR EACH ROW
		EXECUTE FUNCTION %s()
	`, objectDataBacnetSyncFunction)).Error; err != nil {
		return fmt.Errorf("create ObjectData BACnet software-key trigger: %w", err)
	}

	if err := db.Exec(fmt.Sprintf(`
		CREATE OR REPLACE FUNCTION %s()
		RETURNS trigger
		LANGUAGE plpgsql
		AS $$
		BEGIN
			UPDATE object_data_bacnet_objects
			SET
				software_type_normalized = LOWER(BTRIM(NEW.software_type)),
				software_number = NEW.software_number
			WHERE bacnet_object_id = NEW.id;
			RETURN NEW;
		END;
		$$
	`, bacnetObjectTemplateFunction)).Error; err != nil {
		return fmt.Errorf("create BACnet template propagation function: %w", err)
	}
	if err := db.Exec(fmt.Sprintf(`
		DROP TRIGGER IF EXISTS trg_bacnet_object_template_software_keys
		ON bacnet_objects;
		CREATE TRIGGER trg_bacnet_object_template_software_keys
		AFTER UPDATE OF software_type, software_number
		ON bacnet_objects
		FOR EACH ROW
		WHEN (
			OLD.software_type IS DISTINCT FROM NEW.software_type OR
			OLD.software_number IS DISTINCT FROM NEW.software_number
		)
		EXECUTE FUNCTION %s()
	`, bacnetObjectTemplateFunction)).Error; err != nil {
		return fmt.Errorf("create BACnet template propagation trigger: %w", err)
	}

	var orphanCount int64
	if err := db.Raw(`
		SELECT COUNT(*)
		FROM object_data_bacnet_objects AS link
		LEFT JOIN bacnet_objects AS object ON object.id = link.bacnet_object_id
		WHERE object.id IS NULL
	`).Scan(&orphanCount).Error; err != nil {
		return fmt.Errorf("inspect ObjectData BACnet orphan links: %w", err)
	}
	if orphanCount != 0 {
		return fmt.Errorf(
			"cannot install %s: object_data_bacnet_objects contains %d orphan BACnet links",
			objectDataBacnetSoftwareIndex,
			orphanCount,
		)
	}

	if err := db.Exec(`
		UPDATE object_data_bacnet_objects AS link
		SET
			software_type_normalized = LOWER(BTRIM(object.software_type)),
			software_number = object.software_number
		FROM bacnet_objects AS object
		WHERE object.id = link.bacnet_object_id
		  AND (
			link.software_type_normalized IS DISTINCT FROM LOWER(BTRIM(object.software_type)) OR
			link.software_number IS DISTINCT FROM object.software_number
		  )
	`).Error; err != nil {
		return fmt.Errorf("backfill ObjectData BACnet software keys: %w", err)
	}

	var duplicate objectDataBacnetSoftwareDuplicate
	result := db.Raw(`
		SELECT
			object_data_id::text AS object_data_id,
			software_type_normalized,
			software_number,
			array_agg(bacnet_object_id ORDER BY bacnet_object_id)::text AS bacnet_object_ids
		FROM object_data_bacnet_objects
		GROUP BY object_data_id, software_type_normalized, software_number
		HAVING COUNT(*) > 1
		ORDER BY object_data_id, software_type_normalized, software_number
		LIMIT 1
	`).Scan(&duplicate)
	if result.Error != nil {
		return fmt.Errorf("inspect duplicate ObjectData BACnet software keys: %w", result.Error)
	}
	if result.RowsAffected != 0 {
		return fmt.Errorf(
			"cannot install %s: duplicate template software key object_data_id=%s software_type=%q software_number=%d bacnet_object_ids=%s",
			objectDataBacnetSoftwareIndex,
			duplicate.ObjectDataID,
			duplicate.SoftwareTypeNormalized,
			duplicate.SoftwareNumber,
			duplicate.BacnetObjectIDs,
		)
	}

	if err := db.Exec(`
		ALTER TABLE object_data_bacnet_objects
			ALTER COLUMN software_type_normalized SET NOT NULL,
			ALTER COLUMN software_number SET NOT NULL
	`).Error; err != nil {
		return fmt.Errorf("require ObjectData BACnet software keys: %w", err)
	}

	var indexState struct {
		Exists bool `gorm:"column:exists"`
		Valid  bool `gorm:"column:valid"`
	}
	if err := db.Raw(`
		SELECT
			COUNT(*) > 0 AS exists,
			COALESCE(BOOL_AND(index.indisvalid), FALSE) AS valid
		FROM pg_class AS relation
		JOIN pg_index AS index ON index.indexrelid = relation.oid
		WHERE relation.relname = ?
	`, objectDataBacnetSoftwareIndex).Scan(&indexState).Error; err != nil {
		return fmt.Errorf("inspect %s: %w", objectDataBacnetSoftwareIndex, err)
	}
	if indexState.Exists && !indexState.Valid {
		if err := db.Exec(
			"DROP INDEX CONCURRENTLY IF EXISTS " + objectDataBacnetSoftwareIndex,
		).Error; err != nil {
			return fmt.Errorf("drop invalid %s: %w", objectDataBacnetSoftwareIndex, err)
		}
		indexState.Exists = false
	}
	if !indexState.Exists {
		if err := db.Exec(fmt.Sprintf(`
			CREATE UNIQUE INDEX CONCURRENTLY %s
			ON object_data_bacnet_objects (
				object_data_id,
				software_type_normalized,
				software_number
			)
		`, objectDataBacnetSoftwareIndex)).Error; err != nil {
			return fmt.Errorf(
				"create %s; resolve the reported duplicate template keys before retrying: %w",
				objectDataBacnetSoftwareIndex,
				err,
			)
		}
	}
	return nil
}
