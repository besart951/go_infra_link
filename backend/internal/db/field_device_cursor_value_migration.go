package db

import "gorm.io/gorm"

func migrateFieldDeviceCursorValues(database *gorm.DB) error {
	if database.Dialector == nil || database.Dialector.Name() != "postgres" {
		return nil
	}
	for _, statement := range fieldDeviceCursorValueStatements() {
		if err := database.Exec(statement).Error; err != nil {
			return err
		}
	}
	return nil
}

func fieldDeviceCursorValueStatements() []string {
	statements := []string{createFieldDeviceCursorValueTable, refreshFieldDeviceCursorValueFunction}
	statements = append(statements, fieldDeviceCursorValueTriggers()...)
	statements = append(statements, fieldDeviceCursorValueIndexes()...)
	return statements
}

const createFieldDeviceCursorValueTable = `
CREATE TABLE IF NOT EXISTS field_device_cursor_values (
  field_device_id uuid PRIMARY KEY REFERENCES field_devices(id) ON DELETE CASCADE,
  sps_number bigint,
  sps_document_name text,
  specification_supplier text,
  specification_brand text,
  specification_type text,
  additional_info_motor_valve text,
  additional_info_size integer,
  additional_information_installation_location text,
  electrical_connection_ph integer,
  electrical_connection_acdc text,
  electrical_connection_amperage double precision,
  electrical_connection_power double precision,
  electrical_connection_rotation integer
)`

const refreshFieldDeviceCursorValueFunction = `
CREATE OR REPLACE FUNCTION refresh_field_device_cursor_value(target_id uuid) RETURNS void AS $$
BEGIN
  INSERT INTO field_device_cursor_values
  SELECT fd.id, scts.number, scts.document_name,
    specs.specification_supplier, specs.specification_brand, specs.specification_type,
    specs.additional_info_motor_valve, specs.additional_info_size,
    specs.additional_information_installation_location, specs.electrical_connection_ph,
    specs.electrical_connection_acdc, specs.electrical_connection_amperage,
    specs.electrical_connection_power, specs.electrical_connection_rotation
  FROM field_devices fd
  LEFT JOIN sps_controller_system_types scts ON scts.id=fd.sps_controller_system_type_id
  LEFT JOIN specifications specs ON specs.field_device_id=fd.id
  WHERE fd.id=target_id
  ON CONFLICT (field_device_id) DO UPDATE SET
    sps_number=EXCLUDED.sps_number, sps_document_name=EXCLUDED.sps_document_name,
    specification_supplier=EXCLUDED.specification_supplier,
    specification_brand=EXCLUDED.specification_brand, specification_type=EXCLUDED.specification_type,
    additional_info_motor_valve=EXCLUDED.additional_info_motor_valve,
    additional_info_size=EXCLUDED.additional_info_size,
    additional_information_installation_location=EXCLUDED.additional_information_installation_location,
    electrical_connection_ph=EXCLUDED.electrical_connection_ph,
    electrical_connection_acdc=EXCLUDED.electrical_connection_acdc,
    electrical_connection_amperage=EXCLUDED.electrical_connection_amperage,
    electrical_connection_power=EXCLUDED.electrical_connection_power,
    electrical_connection_rotation=EXCLUDED.electrical_connection_rotation;
END;
$$ LANGUAGE plpgsql`

func fieldDeviceCursorValueTriggers() []string {
	return []string{
		backfillFieldDeviceCursorValues,
		`CREATE OR REPLACE FUNCTION sync_field_device_cursor_value() RETURNS trigger AS $$ BEGIN PERFORM refresh_field_device_cursor_value(COALESCE(NEW.id, OLD.id)); RETURN COALESCE(NEW, OLD); END; $$ LANGUAGE plpgsql`,
		`DROP TRIGGER IF EXISTS trg_field_device_cursor_value ON field_devices`,
		`CREATE TRIGGER trg_field_device_cursor_value AFTER INSERT OR UPDATE OF sps_controller_system_type_id ON field_devices FOR EACH ROW EXECUTE FUNCTION sync_field_device_cursor_value()`,
		`CREATE OR REPLACE FUNCTION sync_specification_cursor_value() RETURNS trigger AS $$ BEGIN
  IF TG_OP='UPDATE' AND OLD.field_device_id IS DISTINCT FROM NEW.field_device_id THEN
    PERFORM refresh_field_device_cursor_value(OLD.field_device_id);
  END IF;
  PERFORM refresh_field_device_cursor_value(COALESCE(NEW.field_device_id, OLD.field_device_id));
  RETURN COALESCE(NEW, OLD);
END; $$ LANGUAGE plpgsql`,
		`DROP TRIGGER IF EXISTS trg_specification_cursor_value ON specifications`,
		`CREATE TRIGGER trg_specification_cursor_value AFTER INSERT OR UPDATE OR DELETE ON specifications FOR EACH ROW EXECUTE FUNCTION sync_specification_cursor_value()`,
		`CREATE OR REPLACE FUNCTION sync_sps_cursor_values() RETURNS trigger AS $$ BEGIN UPDATE field_device_cursor_values values SET sps_number=NEW.number, sps_document_name=NEW.document_name FROM field_devices fd WHERE fd.id=values.field_device_id AND fd.sps_controller_system_type_id=NEW.id; RETURN NEW; END; $$ LANGUAGE plpgsql`,
		`DROP TRIGGER IF EXISTS trg_sps_cursor_values ON sps_controller_system_types`,
		`CREATE TRIGGER trg_sps_cursor_values AFTER UPDATE OF number, document_name ON sps_controller_system_types FOR EACH ROW EXECUTE FUNCTION sync_sps_cursor_values()`,
	}
}

const backfillFieldDeviceCursorValues = `
INSERT INTO field_device_cursor_values
SELECT fd.id, scts.number, scts.document_name,
  specs.specification_supplier, specs.specification_brand, specs.specification_type,
  specs.additional_info_motor_valve, specs.additional_info_size,
  specs.additional_information_installation_location, specs.electrical_connection_ph,
  specs.electrical_connection_acdc, specs.electrical_connection_amperage,
  specs.electrical_connection_power, specs.electrical_connection_rotation
FROM field_devices fd
LEFT JOIN sps_controller_system_types scts ON scts.id=fd.sps_controller_system_type_id
LEFT JOIN specifications specs ON specs.field_device_id=fd.id
ON CONFLICT (field_device_id) DO UPDATE SET
  sps_number=EXCLUDED.sps_number, sps_document_name=EXCLUDED.sps_document_name,
  specification_supplier=EXCLUDED.specification_supplier,
  specification_brand=EXCLUDED.specification_brand, specification_type=EXCLUDED.specification_type,
  additional_info_motor_valve=EXCLUDED.additional_info_motor_valve,
  additional_info_size=EXCLUDED.additional_info_size,
  additional_information_installation_location=EXCLUDED.additional_information_installation_location,
  electrical_connection_ph=EXCLUDED.electrical_connection_ph,
  electrical_connection_acdc=EXCLUDED.electrical_connection_acdc,
  electrical_connection_amperage=EXCLUDED.electrical_connection_amperage,
  electrical_connection_power=EXCLUDED.electrical_connection_power,
  electrical_connection_rotation=EXCLUDED.electrical_connection_rotation`

func fieldDeviceCursorValueIndexes() []string {
	indexes := []string{
		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_fd_cursor_values_sps_asc ON field_device_cursor_values ((CASE WHEN sps_number IS NULL THEN 1 ELSE 0 END),sps_number,(CASE WHEN sps_document_name IS NULL THEN 1 ELSE 0 END),sps_document_name,field_device_id)`,
		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_fd_cursor_values_sps_desc ON field_device_cursor_values ((CASE WHEN sps_number IS NULL THEN 1 ELSE 0 END),sps_number DESC,(CASE WHEN sps_document_name IS NULL THEN 1 ELSE 0 END),sps_document_name DESC,field_device_id DESC)`,
		projectionIndex("supplier", "specification_supplier", "ASC"),
		projectionIndex("supplier", "specification_supplier", "DESC"),
	}
	for _, column := range specificationCursorColumns() {
		indexes = append(indexes, projectionIndex(compactIndexColumn(column), column, "ASC"))
		indexes = append(indexes, projectionIndex(compactIndexColumn(column), column, "DESC"))
	}
	return indexes
}

func projectionIndex(name, column, direction string) string {
	suffix := "asc"
	if direction == "DESC" {
		suffix = "desc"
	}
	return "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_fd_cursor_values_" + name + "_" + suffix +
		" ON field_device_cursor_values ((CASE WHEN " + column + " IS NULL THEN 1 ELSE 0 END)," +
		column + " " + direction + ",field_device_id " + direction + ")"
}
