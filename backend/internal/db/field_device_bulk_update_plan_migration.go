package db

import "gorm.io/gorm"

func migrateFieldDeviceBulkUpdatePlans(database *gorm.DB) error {
	if database.Dialector == nil || database.Dialector.Name() != "postgres" {
		return nil
	}
	return database.Exec(`
CREATE TABLE IF NOT EXISTS field_device_bulk_update_plan_items (
  owner_id uuid NOT NULL,
  job_id uuid NOT NULL,
  ordinal bigint NOT NULL,
  group_ordinal bigint NOT NULL,
  dependency_group_id uuid NOT NULL,
  field_device_id uuid NOT NULL,
  command jsonb NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (owner_id,job_id,ordinal),
  FOREIGN KEY (owner_id,job_id) REFERENCES facility_jobs(owner_id,id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_fd_bulk_plan_group
  ON field_device_bulk_update_plan_items (owner_id,job_id,group_ordinal,ordinal);
CREATE INDEX IF NOT EXISTS idx_fd_bulk_plan_current_id
  ON field_device_bulk_update_plan_items (field_device_id,job_id);`).Error
}
