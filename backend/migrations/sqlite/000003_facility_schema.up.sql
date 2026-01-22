PRAGMA foreign_keys = ON;

-- SQLite facility schema (dev)
-- Notes:
-- - UUIDs stored as TEXT
-- - TIMESTAMPTZ stored as TEXT (ISO8601 recommended)
-- - JSON stored as TEXT

CREATE TABLE IF NOT EXISTS system_types (
    id TEXT PRIMARY KEY,
    created_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP),
    updated_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP),
    deleted_at TEXT,
    number_min INTEGER NOT NULL,
    number_max INTEGER NOT NULL,
    name TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_system_types_deleted_at ON system_types(deleted_at);

CREATE TABLE IF NOT EXISTS system_parts (
    id TEXT PRIMARY KEY,
    created_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP),
    updated_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP),
    deleted_at TEXT,
    short_name TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_system_parts_short_name ON system_parts(short_name);
CREATE UNIQUE INDEX IF NOT EXISTS idx_system_parts_name ON system_parts(name);
CREATE INDEX IF NOT EXISTS idx_system_parts_deleted_at ON system_parts(deleted_at);

-- Specifications (final shape as of migration 000005)
CREATE TABLE IF NOT EXISTS specifications (
    id TEXT PRIMARY KEY,
    created_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP),
    updated_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP),
    deleted_at TEXT,
    specification_supplier TEXT,
    specification_brand TEXT,
    specification_type TEXT,
    additional_info_motor_valve TEXT,
    additional_info_size INTEGER,
    additional_information_installation_location TEXT,
    electrical_connection_ph INTEGER,
    electrical_connection_acdc TEXT,
    electrical_connection_amperage REAL,
    electrical_connection_power REAL,
    electrical_connection_rotation INTEGER,
    field_device_id TEXT,
    FOREIGN KEY (field_device_id) REFERENCES field_devices(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_specifications_deleted_at ON specifications(deleted_at);
-- 1:1 for active specifications
CREATE UNIQUE INDEX IF NOT EXISTS idx_specifications_field_device_id
    ON specifications(field_device_id)
    WHERE deleted_at IS NULL AND field_device_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS notification_classes (
    id TEXT PRIMARY KEY,
    created_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP),
    updated_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP),
    deleted_at TEXT,
    event_category TEXT NOT NULL,
    nc INTEGER NOT NULL,
    object_description TEXT NOT NULL,
    internal_description TEXT NOT NULL,
    meaning TEXT NOT NULL,
    ack_required_not_normal INTEGER NOT NULL DEFAULT 0,
    ack_required_error INTEGER NOT NULL DEFAULT 0,
    ack_required_normal INTEGER NOT NULL DEFAULT 0,
    norm_not_normal INTEGER NOT NULL,
    norm_error INTEGER NOT NULL,
    norm_normal INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_notification_classes_deleted_at ON notification_classes(deleted_at);

CREATE TABLE IF NOT EXISTS alarm_definitions (
    id TEXT PRIMARY KEY,
    created_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP),
    updated_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP),
    deleted_at TEXT,
    name TEXT NOT NULL,
    alarm_note TEXT
);

CREATE INDEX IF NOT EXISTS idx_alarm_definitions_deleted_at ON alarm_definitions(deleted_at);

CREATE TABLE IF NOT EXISTS state_texts (
    id TEXT PRIMARY KEY,
    created_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP),
    updated_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP),
    deleted_at TEXT,
    ref_number INTEGER NOT NULL,
    state_text1 TEXT,
    state_text2 TEXT,
    state_text3 TEXT,
    state_text4 TEXT,
    state_text5 TEXT,
    state_text6 TEXT,
    state_text7 TEXT,
    state_text8 TEXT,
    state_text9 TEXT,
    state_text10 TEXT,
    state_text11 TEXT,
    state_text12 TEXT,
    state_text13 TEXT,
    state_text14 TEXT,
    state_text15 TEXT,
    state_text16 TEXT
);

CREATE INDEX IF NOT EXISTS idx_state_texts_deleted_at ON state_texts(deleted_at);

CREATE TABLE IF NOT EXISTS apparats (
    id TEXT PRIMARY KEY,
    created_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP),
    updated_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP),
    deleted_at TEXT,
    short_name TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_apparat_composite ON apparats(short_name, name);
CREATE INDEX IF NOT EXISTS idx_apparats_deleted_at ON apparats(deleted_at);

CREATE TABLE IF NOT EXISTS buildings (
    id TEXT PRIMARY KEY,
    created_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP),
    updated_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP),
    deleted_at TEXT,
    iws_code TEXT,
    building_group INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_buildings_deleted_at ON buildings(deleted_at);
CREATE INDEX IF NOT EXISTS idx_buildings_iws_code ON buildings(iws_code);

CREATE TABLE IF NOT EXISTS control_cabinets (
    id TEXT PRIMARY KEY,
    created_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP),
    updated_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP),
    deleted_at TEXT,
    building_id TEXT NOT NULL,
    project_id TEXT,
    control_cabinet_nr TEXT,
    FOREIGN KEY (building_id) REFERENCES buildings(id) ON DELETE CASCADE,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_control_cabinets_deleted_at ON control_cabinets(deleted_at);
CREATE INDEX IF NOT EXISTS idx_control_cabinets_building_id ON control_cabinets(building_id);
CREATE INDEX IF NOT EXISTS idx_control_cabinets_project_id ON control_cabinets(project_id);
CREATE INDEX IF NOT EXISTS idx_control_cabinets_control_cabinet_nr ON control_cabinets(control_cabinet_nr);

CREATE TABLE IF NOT EXISTS sps_controllers (
    id TEXT PRIMARY KEY,
    created_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP),
    updated_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP),
    deleted_at TEXT,
    control_cabinet_id TEXT NOT NULL,
    project_id TEXT,
    ga_device TEXT,
    device_name TEXT NOT NULL,
    device_description TEXT,
    device_location TEXT,
    ip_address TEXT,
    subnet TEXT,
    gateway TEXT,
    vlan TEXT,
    FOREIGN KEY (control_cabinet_id) REFERENCES control_cabinets(id) ON DELETE CASCADE,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_sps_controllers_deleted_at ON sps_controllers(deleted_at);
CREATE INDEX IF NOT EXISTS idx_sps_controllers_control_cabinet_id ON sps_controllers(control_cabinet_id);
CREATE INDEX IF NOT EXISTS idx_sps_controllers_project_id ON sps_controllers(project_id);
CREATE INDEX IF NOT EXISTS idx_sps_controllers_ip_address ON sps_controllers(ip_address);
CREATE INDEX IF NOT EXISTS idx_sps_controllers_vlan ON sps_controllers(vlan);

CREATE TABLE IF NOT EXISTS sps_controller_system_types (
    id TEXT PRIMARY KEY,
    created_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP),
    updated_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP),
    deleted_at TEXT,
    number INTEGER,
    document_name TEXT,
    sps_controller_id TEXT NOT NULL,
    system_type_id TEXT NOT NULL,
    FOREIGN KEY (sps_controller_id) REFERENCES sps_controllers(id) ON DELETE CASCADE,
    FOREIGN KEY (system_type_id) REFERENCES system_types(id) ON DELETE RESTRICT
);

CREATE INDEX IF NOT EXISTS idx_sps_controller_system_types_deleted_at ON sps_controller_system_types(deleted_at);
CREATE INDEX IF NOT EXISTS idx_sps_controller_system_types_sps_controller_id ON sps_controller_system_types(sps_controller_id);
CREATE INDEX IF NOT EXISTS idx_sps_controller_system_types_system_type_id ON sps_controller_system_types(system_type_id);

-- Field devices (final shape as of migrations 000004 + 000005)
CREATE TABLE IF NOT EXISTS field_devices (
    id TEXT PRIMARY KEY,
    created_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP),
    updated_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP),
    deleted_at TEXT,
    bmk TEXT,
    description TEXT,
    apparat_nr INTEGER NOT NULL,
    sps_controller_system_type_id TEXT NOT NULL,
    system_part_id TEXT,
    project_id TEXT,
    apparat_id TEXT NOT NULL,
    CHECK (apparat_nr BETWEEN 1 AND 99),
    FOREIGN KEY (sps_controller_system_type_id) REFERENCES sps_controller_system_types(id) ON DELETE RESTRICT,
    FOREIGN KEY (system_part_id) REFERENCES system_parts(id) ON DELETE CASCADE,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE SET NULL,
    FOREIGN KEY (apparat_id) REFERENCES apparats(id) ON DELETE RESTRICT
);

CREATE INDEX IF NOT EXISTS idx_field_devices_deleted_at ON field_devices(deleted_at);
CREATE INDEX IF NOT EXISTS idx_field_devices_bmk ON field_devices(bmk);
CREATE INDEX IF NOT EXISTS idx_field_devices_apparat_nr ON field_devices(apparat_nr);
CREATE INDEX IF NOT EXISTS idx_field_devices_sps_controller_system_type_id ON field_devices(sps_controller_system_type_id);
CREATE INDEX IF NOT EXISTS idx_field_devices_project_id ON field_devices(project_id);

-- Unique apparat_nr per (sps_controller_system_type, system_part, apparat) for active rows
CREATE UNIQUE INDEX IF NOT EXISTS idx_field_devices_unique_apparat_nr
    ON field_devices (
        sps_controller_system_type_id,
        apparat_id,
        COALESCE(system_part_id, '00000000-0000-0000-0000-000000000000'),
        apparat_nr
    )
    WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS bacnet_objects (
    id TEXT PRIMARY KEY,
    created_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP),
    updated_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP),
    deleted_at TEXT,
    text_fix TEXT NOT NULL,
    description TEXT,
    gms_visible INTEGER NOT NULL DEFAULT 0,
    optional INTEGER NOT NULL DEFAULT 0,
    text_individual TEXT,
    software_type TEXT NOT NULL,
    software_number INTEGER NOT NULL,
    hardware_type TEXT NOT NULL,
    hardware_quantity INTEGER NOT NULL DEFAULT 1,
    field_device_id TEXT,
    software_reference_id TEXT,
    state_text_id TEXT,
    notification_class_id TEXT,
    alarm_definition_id TEXT,
    FOREIGN KEY (field_device_id) REFERENCES field_devices(id) ON DELETE CASCADE,
    FOREIGN KEY (software_reference_id) REFERENCES bacnet_objects(id) ON DELETE SET NULL,
    FOREIGN KEY (state_text_id) REFERENCES state_texts(id) ON DELETE SET NULL,
    FOREIGN KEY (notification_class_id) REFERENCES notification_classes(id) ON DELETE SET NULL,
    FOREIGN KEY (alarm_definition_id) REFERENCES alarm_definitions(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_bacnet_objects_deleted_at ON bacnet_objects(deleted_at);
CREATE UNIQUE INDEX IF NOT EXISTS idx_bacnet_fd_text ON bacnet_objects(field_device_id, text_fix);

CREATE TABLE IF NOT EXISTS object_data (
    id TEXT PRIMARY KEY,
    created_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP),
    updated_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP),
    deleted_at TEXT,
    description TEXT NOT NULL,
    version TEXT NOT NULL DEFAULT '1.0.0',
    is_active INTEGER NOT NULL DEFAULT 1,
    project_id TEXT,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_object_data_deleted_at ON object_data(deleted_at);
CREATE INDEX IF NOT EXISTS idx_object_data_is_active ON object_data(is_active);
CREATE UNIQUE INDEX IF NOT EXISTS idx_obj_data_proj_desc ON object_data(project_id, description);

CREATE TABLE IF NOT EXISTS apparat_system_parts (
    apparat_id TEXT NOT NULL,
    system_part_id TEXT NOT NULL,
    PRIMARY KEY (apparat_id, system_part_id),
    FOREIGN KEY (apparat_id) REFERENCES apparats(id) ON DELETE CASCADE,
    FOREIGN KEY (system_part_id) REFERENCES system_parts(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_apparat_system_parts_system_part_id ON apparat_system_parts(system_part_id);

CREATE TABLE IF NOT EXISTS object_data_apparats (
    object_data_id TEXT NOT NULL,
    apparat_id TEXT NOT NULL,
    PRIMARY KEY (object_data_id, apparat_id),
    FOREIGN KEY (object_data_id) REFERENCES object_data(id) ON DELETE CASCADE,
    FOREIGN KEY (apparat_id) REFERENCES apparats(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_object_data_apparats_apparat_id ON object_data_apparats(apparat_id);

CREATE TABLE IF NOT EXISTS object_data_bacnet_objects (
    object_data_id TEXT NOT NULL,
    bacnet_object_id TEXT NOT NULL,
    PRIMARY KEY (object_data_id, bacnet_object_id),
    FOREIGN KEY (object_data_id) REFERENCES object_data(id) ON DELETE CASCADE,
    FOREIGN KEY (bacnet_object_id) REFERENCES bacnet_objects(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_object_data_bacnet_objects_bacnet_object_id ON object_data_bacnet_objects(bacnet_object_id);

CREATE TABLE IF NOT EXISTS object_data_histories (
    id TEXT PRIMARY KEY,
    created_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP),
    updated_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP),
    deleted_at TEXT,
    object_data_id TEXT,
    user_id TEXT,
    action TEXT NOT NULL,
    changes TEXT,
    FOREIGN KEY (object_data_id) REFERENCES object_data(id) ON DELETE SET NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_object_data_histories_deleted_at ON object_data_histories(deleted_at);
