package db

import "gorm.io/gorm"

func migrateProjectFieldDeviceCursorProjection(database *gorm.DB) error {
	if database.Dialector == nil || database.Dialector.Name() != "postgres" {
		return nil
	}
	for _, statement := range projectFieldDeviceCursorStatements() {
		if err := database.Exec(statement).Error; err != nil {
			return err
		}
	}
	return nil
}

func projectFieldDeviceCursorStatements() []string {
	return []string{
		`ALTER TABLE project_field_devices ADD COLUMN IF NOT EXISTS field_device_created_at timestamptz`,
		`UPDATE project_field_devices links SET field_device_created_at=devices.created_at FROM field_devices devices WHERE devices.id=links.field_device_id AND links.field_device_created_at IS NULL`,
		`ALTER TABLE project_field_devices ALTER COLUMN field_device_created_at SET NOT NULL`,
		projectFieldDeviceCursorTriggerFunction,
		`DROP TRIGGER IF EXISTS trg_project_field_device_cursor ON project_field_devices`,
		`CREATE TRIGGER trg_project_field_device_cursor BEFORE INSERT OR UPDATE OF field_device_id ON project_field_devices FOR EACH ROW EXECUTE FUNCTION sync_project_field_device_cursor()`,
		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_pfd_project_created_fd_asc ON project_field_devices (project_id,field_device_created_at,field_device_id)`,
		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_pfd_project_created_fd_desc ON project_field_devices (project_id,field_device_created_at DESC,field_device_id DESC)`,
	}
}

const projectFieldDeviceCursorTriggerFunction = `
CREATE OR REPLACE FUNCTION sync_project_field_device_cursor() RETURNS trigger AS $$
BEGIN
  IF TG_OP='INSERT' AND NEW.field_device_created_at IS NOT NULL THEN
    RETURN NEW;
  END IF;
  IF TG_OP='UPDATE' AND NEW.field_device_id IS NOT DISTINCT FROM OLD.field_device_id AND NEW.field_device_created_at IS NOT NULL THEN
    RETURN NEW;
  END IF;
  SELECT created_at INTO NEW.field_device_created_at FROM field_devices WHERE id=NEW.field_device_id;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql`
