-- Field devices constraints (performance + integrity)

-- apparat_nr must be present and between 1 and 99
ALTER TABLE field_devices
    ALTER COLUMN apparat_nr SET NOT NULL;

ALTER TABLE field_devices
    ADD CONSTRAINT chk_field_devices_apparat_nr_range CHECK (apparat_nr BETWEEN 1 AND 99);

-- apparat_nr must be unique per (sps_controller_system_type, system_part, apparat)
-- Note: system_part_id is nullable, so COALESCE is used to treat NULL as a single bucket.
CREATE UNIQUE INDEX idx_field_devices_unique_apparat_nr
    ON field_devices (
        sps_controller_system_type_id,
        apparat_id,
        COALESCE(system_part_id, '00000000-0000-0000-0000-000000000000'::uuid),
        apparat_nr
    )
    WHERE deleted_at IS NULL;
