-- Revert FieldDevice <-> Specification relationship back to field_devices.specification_id

ALTER TABLE field_devices
    ADD COLUMN IF NOT EXISTS specification_id UUID;

-- Backfill from specifications.field_device_id
UPDATE field_devices fd
SET specification_id = s.id
FROM specifications s
WHERE s.field_device_id = fd.id
  AND s.deleted_at IS NULL
  AND fd.deleted_at IS NULL
  AND fd.specification_id IS NULL;

ALTER TABLE field_devices
    ADD CONSTRAINT fk_field_devices_specification
        FOREIGN KEY (specification_id) REFERENCES specifications(id) ON DELETE CASCADE;

DROP INDEX IF EXISTS idx_specifications_field_device_id;

ALTER TABLE specifications
    DROP CONSTRAINT IF EXISTS fk_specifications_field_device;

ALTER TABLE specifications
    DROP COLUMN IF EXISTS field_device_id;
