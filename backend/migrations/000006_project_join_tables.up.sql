-- Refactor project relationships to use join tables
-- This migration removes direct project_id foreign keys and creates many-to-many relationships

-- Create join tables
CREATE TABLE project_control_cabinets (
    project_id UUID NOT NULL,
    control_cabinet_id UUID NOT NULL,
    PRIMARY KEY (project_id, control_cabinet_id),
    CONSTRAINT fk_project_control_cabinets_project FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    CONSTRAINT fk_project_control_cabinets_control_cabinet FOREIGN KEY (control_cabinet_id) REFERENCES control_cabinets(id) ON DELETE CASCADE
);

CREATE INDEX idx_project_control_cabinets_control_cabinet_id ON project_control_cabinets(control_cabinet_id);

CREATE TABLE project_sps_controllers (
    project_id UUID NOT NULL,
    sps_controller_id UUID NOT NULL,
    PRIMARY KEY (project_id, sps_controller_id),
    CONSTRAINT fk_project_sps_controllers_project FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    CONSTRAINT fk_project_sps_controllers_sps_controller FOREIGN KEY (sps_controller_id) REFERENCES sps_controllers(id) ON DELETE CASCADE
);

CREATE INDEX idx_project_sps_controllers_sps_controller_id ON project_sps_controllers(sps_controller_id);

CREATE TABLE project_field_devices (
    project_id UUID NOT NULL,
    field_device_id UUID NOT NULL,
    PRIMARY KEY (project_id, field_device_id),
    CONSTRAINT fk_project_field_devices_project FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    CONSTRAINT fk_project_field_devices_field_device FOREIGN KEY (field_device_id) REFERENCES field_devices(id) ON DELETE CASCADE
);

CREATE INDEX idx_project_field_devices_field_device_id ON project_field_devices(field_device_id);

-- Migrate existing data to join tables
INSERT INTO project_control_cabinets (project_id, control_cabinet_id)
SELECT project_id, id FROM control_cabinets WHERE project_id IS NOT NULL AND deleted_at IS NULL;

INSERT INTO project_sps_controllers (project_id, sps_controller_id)
SELECT project_id, id FROM sps_controllers WHERE project_id IS NOT NULL AND deleted_at IS NULL;

INSERT INTO project_field_devices (project_id, field_device_id)
SELECT project_id, id FROM field_devices WHERE project_id IS NOT NULL AND deleted_at IS NULL;

-- Remove old foreign key constraints and columns
ALTER TABLE control_cabinets DROP CONSTRAINT IF EXISTS fk_control_cabinets_project;
ALTER TABLE control_cabinets DROP COLUMN IF EXISTS project_id;
DROP INDEX IF EXISTS idx_control_cabinets_project_id;

ALTER TABLE sps_controllers DROP CONSTRAINT IF EXISTS fk_sps_controllers_project;
ALTER TABLE sps_controllers DROP COLUMN IF EXISTS project_id;
DROP INDEX IF EXISTS idx_sps_controllers_project_id;

ALTER TABLE field_devices DROP CONSTRAINT IF EXISTS fk_field_devices_project;
ALTER TABLE field_devices DROP COLUMN IF EXISTS project_id;
DROP INDEX IF EXISTS idx_field_devices_project_id;

-- Add unique constraints and checks

-- system_types.name must be unique
CREATE UNIQUE INDEX IF NOT EXISTS idx_system_types_name_unique ON system_types(name) WHERE deleted_at IS NULL;

-- system_types number ranges must not overlap
-- This is enforced by an exclusion constraint using range types
CREATE EXTENSION IF NOT EXISTS btree_gist;
ALTER TABLE system_types ADD CONSTRAINT chk_system_types_no_overlap 
    EXCLUDE USING GIST (int4range(number_min, number_max, '[]') WITH &&) WHERE (deleted_at IS NULL);

-- system_parts.name or short_name must be unique (already has unique constraints, ensure they exist)
-- Already defined in 000003_facility_schema.up.sql

-- apparats.name or short_name must be unique (already has composite unique)
-- Already defined in 000003_facility_schema.up.sql

-- buildings.iws_code + building_group must be unique
CREATE UNIQUE INDEX IF NOT EXISTS idx_buildings_iws_code_group_unique ON buildings(iws_code, building_group) WHERE deleted_at IS NULL;

-- control_cabinets.control_cabinet_nr must be unique per building
CREATE UNIQUE INDEX IF NOT EXISTS idx_control_cabinet_nr_per_building ON control_cabinets(building_id, control_cabinet_nr) WHERE deleted_at IS NULL AND control_cabinet_nr IS NOT NULL;

-- field_devices: system_part_id must be NOT NULL
ALTER TABLE field_devices ALTER COLUMN system_part_id SET NOT NULL;

-- Update the unique constraint for field_devices to use system_part_id (not COALESCE)
-- Drop old index that used COALESCE
DROP INDEX IF EXISTS idx_field_devices_unique_apparat_nr;

-- Create new index without COALESCE since system_part_id is now NOT NULL
CREATE UNIQUE INDEX idx_field_devices_unique_apparat_nr
    ON field_devices (
        sps_controller_system_type_id,
        system_part_id,
        apparat_id,
        apparat_nr
    )
    WHERE deleted_at IS NULL;
