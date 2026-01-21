-- FieldDevice <-> Specification: enforce 1:1 (nullable on FieldDevice)
--
-- New model:
--   - field_devices no longer stores specification_id
--   - specifications stores field_device_id (unique, soft-delete aware)
--
-- Note: field_device_id is kept nullable to allow existing orphan specifications
-- to remain until cleaned up, but the application requires it for new rows.

ALTER TABLE specifications
    ADD COLUMN IF NOT EXISTS field_device_id UUID;

-- Backfill links from previous schema if present
UPDATE specifications s
SET field_device_id = fd.id
FROM field_devices fd
WHERE fd.specification_id = s.id
  AND fd.deleted_at IS NULL
  AND s.deleted_at IS NULL
  AND s.field_device_id IS NULL;

-- FK from specification to field_device (hard-delete cascade)
ALTER TABLE specifications
    ADD CONSTRAINT fk_specifications_field_device
        FOREIGN KEY (field_device_id) REFERENCES field_devices(id) ON DELETE CASCADE;

-- Enforce 1:1 for active specifications
CREATE UNIQUE INDEX IF NOT EXISTS idx_specifications_field_device_id
    ON specifications(field_device_id)
    WHERE deleted_at IS NULL AND field_device_id IS NOT NULL;

-- Drop old FK/column (field_devices -> specifications)
ALTER TABLE field_devices
    DROP CONSTRAINT IF EXISTS fk_field_devices_specification;

ALTER TABLE field_devices
    DROP COLUMN IF EXISTS specification_id;
