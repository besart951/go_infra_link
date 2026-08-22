package db

import "gorm.io/gorm"

func migrateFieldDeviceBuildingCursorProjection(database *gorm.DB) error {
	if database.Dialector == nil || database.Dialector.Name() != "postgres" {
		return nil
	}
	for _, statement := range fieldDeviceBuildingCursorStatements() {
		if err := database.Exec(statement).Error; err != nil {
			return err
		}
	}
	return nil
}

func fieldDeviceBuildingCursorStatements() []string {
	return []string{
		createFieldDeviceBuildingCursorTable, backfillFieldDeviceBuildingCursor,
		refreshFieldDeviceBuildingCursorFunction,
		`DROP TRIGGER IF EXISTS trg_field_device_building_cursor ON field_devices`,
		`CREATE TRIGGER trg_field_device_building_cursor AFTER INSERT OR UPDATE OF sps_controller_system_type_id ON field_devices FOR EACH ROW EXECUTE FUNCTION sync_field_device_building_cursor()`,
		refreshSystemTypeBuildingCursorsFunction,
		`DROP TRIGGER IF EXISTS trg_system_type_building_cursors ON sps_controller_system_types`,
		`CREATE TRIGGER trg_system_type_building_cursors AFTER UPDATE OF sps_controller_id ON sps_controller_system_types FOR EACH ROW EXECUTE FUNCTION sync_system_type_building_cursors()`,
		refreshControllerBuildingCursorsFunction,
		`DROP TRIGGER IF EXISTS trg_controller_building_cursors ON sps_controllers`,
		`CREATE TRIGGER trg_controller_building_cursors AFTER UPDATE OF control_cabinet_id ON sps_controllers FOR EACH ROW EXECUTE FUNCTION sync_controller_building_cursors()`,
		refreshCabinetBuildingCursorsFunction,
		`DROP TRIGGER IF EXISTS trg_cabinet_building_cursors ON control_cabinets`,
		`CREATE TRIGGER trg_cabinet_building_cursors AFTER UPDATE OF building_id ON control_cabinets FOR EACH ROW EXECUTE FUNCTION sync_cabinet_building_cursors()`,
		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_fd_building_cursor_asc ON field_device_building_cursor_values (building_id,field_device_created_at,field_device_id)`,
		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_fd_building_cursor_desc ON field_device_building_cursor_values (building_id,field_device_created_at DESC,field_device_id DESC)`,
	}
}

const createFieldDeviceBuildingCursorTable = `
CREATE TABLE IF NOT EXISTS field_device_building_cursor_values (
  field_device_id uuid PRIMARY KEY REFERENCES field_devices(id) ON DELETE CASCADE,
  building_id uuid NOT NULL REFERENCES buildings(id),
  field_device_created_at timestamptz NOT NULL
)`

const backfillFieldDeviceBuildingCursor = `
INSERT INTO field_device_building_cursor_values (field_device_id,building_id,field_device_created_at)
SELECT device.id,cabinet.building_id,device.created_at
FROM field_devices device
JOIN sps_controller_system_types system_type ON system_type.id=device.sps_controller_system_type_id
JOIN sps_controllers controller ON controller.id=system_type.sps_controller_id
JOIN control_cabinets cabinet ON cabinet.id=controller.control_cabinet_id
ON CONFLICT (field_device_id) DO UPDATE SET
  building_id=EXCLUDED.building_id,field_device_created_at=EXCLUDED.field_device_created_at`

const refreshFieldDeviceBuildingCursorFunction = `
CREATE OR REPLACE FUNCTION sync_field_device_building_cursor() RETURNS trigger AS $$
BEGIN
  INSERT INTO field_device_building_cursor_values (field_device_id,building_id,field_device_created_at)
  SELECT NEW.id,cabinet.building_id,NEW.created_at
  FROM sps_controller_system_types system_type
  JOIN sps_controllers controller ON controller.id=system_type.sps_controller_id
  JOIN control_cabinets cabinet ON cabinet.id=controller.control_cabinet_id
  WHERE system_type.id=NEW.sps_controller_system_type_id
  ON CONFLICT (field_device_id) DO UPDATE SET
    building_id=EXCLUDED.building_id,field_device_created_at=EXCLUDED.field_device_created_at;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql`

const refreshSystemTypeBuildingCursorsFunction = `
CREATE OR REPLACE FUNCTION sync_system_type_building_cursors() RETURNS trigger AS $$
BEGIN
  UPDATE field_device_building_cursor_values values SET building_id=cabinet.building_id
  FROM field_devices device
  JOIN sps_controllers controller ON controller.id=NEW.sps_controller_id
  JOIN control_cabinets cabinet ON cabinet.id=controller.control_cabinet_id
  WHERE device.id=values.field_device_id AND device.sps_controller_system_type_id=NEW.id;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql`

const refreshControllerBuildingCursorsFunction = `
CREATE OR REPLACE FUNCTION sync_controller_building_cursors() RETURNS trigger AS $$
BEGIN
  UPDATE field_device_building_cursor_values values SET building_id=cabinet.building_id
  FROM field_devices device
  JOIN sps_controller_system_types system_type ON system_type.id=device.sps_controller_system_type_id
  JOIN control_cabinets cabinet ON cabinet.id=NEW.control_cabinet_id
  WHERE device.id=values.field_device_id AND system_type.sps_controller_id=NEW.id;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql`

const refreshCabinetBuildingCursorsFunction = `
CREATE OR REPLACE FUNCTION sync_cabinet_building_cursors() RETURNS trigger AS $$
BEGIN
  UPDATE field_device_building_cursor_values values SET building_id=NEW.building_id
  FROM field_devices device
  JOIN sps_controller_system_types system_type ON system_type.id=device.sps_controller_system_type_id
  JOIN sps_controllers controller ON controller.id=system_type.sps_controller_id
  WHERE device.id=values.field_device_id AND controller.control_cabinet_id=NEW.id;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql`
