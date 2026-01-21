-- Revert field devices constraints

DROP INDEX IF EXISTS idx_field_devices_unique_apparat_nr;

ALTER TABLE field_devices
    DROP CONSTRAINT IF EXISTS chk_field_devices_apparat_nr_range;

ALTER TABLE field_devices
    ALTER COLUMN apparat_nr DROP NOT NULL;
