-- Rollback project join tables migration

-- Restore old unique index for field_devices with COALESCE
DROP INDEX IF EXISTS idx_field_devices_unique_apparat_nr;
CREATE UNIQUE INDEX idx_field_devices_unique_apparat_nr
    ON field_devices (
        sps_controller_system_type_id,
        apparat_id,
        COALESCE(system_part_id, '00000000-0000-0000-0000-000000000000'::uuid),
        apparat_nr
    )
    WHERE deleted_at IS NULL;

-- Make system_part_id nullable again
ALTER TABLE field_devices ALTER COLUMN system_part_id DROP NOT NULL;

-- Drop new unique indexes and constraints
DROP INDEX IF EXISTS idx_control_cabinet_nr_per_building;
DROP INDEX IF EXISTS idx_buildings_iws_code_group_unique;
ALTER TABLE system_types DROP CONSTRAINT IF EXISTS chk_system_types_no_overlap;
DROP INDEX IF EXISTS idx_system_types_name_unique;

-- Restore project_id columns to original tables
ALTER TABLE control_cabinets ADD COLUMN project_id UUID;
ALTER TABLE sps_controllers ADD COLUMN project_id UUID;
ALTER TABLE field_devices ADD COLUMN project_id UUID;

-- Restore data from join tables
UPDATE control_cabinets
SET project_id = pcc.project_id
FROM project_control_cabinets pcc
WHERE control_cabinets.id = pcc.control_cabinet_id;

UPDATE sps_controllers
SET project_id = psc.project_id
FROM project_sps_controllers psc
WHERE sps_controllers.id = psc.sps_controller_id;

UPDATE field_devices
SET project_id = pfd.project_id
FROM project_field_devices pfd
WHERE field_devices.id = pfd.field_device_id;

-- Restore foreign key constraints
ALTER TABLE control_cabinets
    ADD CONSTRAINT fk_control_cabinets_project FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE SET NULL;
CREATE INDEX idx_control_cabinets_project_id ON control_cabinets(project_id);

ALTER TABLE sps_controllers
    ADD CONSTRAINT fk_sps_controllers_project FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE SET NULL;
CREATE INDEX idx_sps_controllers_project_id ON sps_controllers(project_id);

ALTER TABLE field_devices
    ADD CONSTRAINT fk_field_devices_project FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE SET NULL;
CREATE INDEX idx_field_devices_project_id ON field_devices(project_id);

-- Drop join tables
DROP TABLE IF EXISTS project_field_devices;
DROP TABLE IF EXISTS project_sps_controllers;
DROP TABLE IF EXISTS project_control_cabinets;
