-- Facility schema migration

-- System types
CREATE TABLE system_types (
    id UUID PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    number_min INT NOT NULL,
    number_max INT NOT NULL,
    name VARCHAR(150) NOT NULL
);

CREATE INDEX idx_system_types_deleted_at ON system_types(deleted_at);

-- System parts
CREATE TABLE system_parts (
    id UUID PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    short_name VARCHAR(10) NOT NULL,
    name VARCHAR(250) NOT NULL,
    description VARCHAR(250)
);

CREATE UNIQUE INDEX idx_system_parts_short_name ON system_parts(short_name);
CREATE UNIQUE INDEX idx_system_parts_name ON system_parts(name);
CREATE INDEX idx_system_parts_deleted_at ON system_parts(deleted_at);

-- Specifications
CREATE TABLE specifications (
    id UUID PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    specification_supplier VARCHAR(250),
    specification_brand VARCHAR(250),
    specification_type VARCHAR(250),
    additional_info_motor_valve VARCHAR(250),
    additional_info_size INT,
    additional_information_installation_location VARCHAR(250),
    electrical_connection_ph INT,
    electrical_connection_acdc VARCHAR(2),
    electrical_connection_amperage DOUBLE PRECISION,
    electrical_connection_power DOUBLE PRECISION,
    electrical_connection_rotation INT
);

CREATE INDEX idx_specifications_deleted_at ON specifications(deleted_at);

-- Notification classes
CREATE TABLE notification_classes (
    id UUID PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    event_category VARCHAR(50) NOT NULL,
    nc INT NOT NULL,
    object_description VARCHAR(250) NOT NULL,
    internal_description VARCHAR(250) NOT NULL,
    meaning VARCHAR(250) NOT NULL,
    ack_required_not_normal BOOLEAN NOT NULL DEFAULT FALSE,
    ack_required_error BOOLEAN NOT NULL DEFAULT FALSE,
    ack_required_normal BOOLEAN NOT NULL DEFAULT FALSE,
    norm_not_normal INT NOT NULL,
    norm_error INT NOT NULL,
    norm_normal INT NOT NULL
);

CREATE INDEX idx_notification_classes_deleted_at ON notification_classes(deleted_at);

-- Alarm definitions
CREATE TABLE alarm_definitions (
    id UUID PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    name VARCHAR(350) NOT NULL,
    alarm_note VARCHAR(250)
);

CREATE INDEX idx_alarm_definitions_deleted_at ON alarm_definitions(deleted_at);

-- State texts
CREATE TABLE state_texts (
    id UUID PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    ref_number INT NOT NULL,
    state_text1 VARCHAR(100),
    state_text2 VARCHAR(100),
    state_text3 VARCHAR(100),
    state_text4 VARCHAR(100),
    state_text5 VARCHAR(100),
    state_text6 VARCHAR(100),
    state_text7 VARCHAR(100),
    state_text8 VARCHAR(100),
    state_text9 VARCHAR(100),
    state_text10 VARCHAR(100),
    state_text11 VARCHAR(100),
    state_text12 VARCHAR(100),
    state_text13 VARCHAR(100),
    state_text14 VARCHAR(100),
    state_text15 VARCHAR(100),
    state_text16 VARCHAR(100)
);

CREATE INDEX idx_state_texts_deleted_at ON state_texts(deleted_at);

-- Apparats
CREATE TABLE apparats (
    id UUID PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    short_name TEXT NOT NULL,
    name VARCHAR(250) NOT NULL,
    description VARCHAR(250)
);

CREATE UNIQUE INDEX idx_apparat_composite ON apparats(short_name, name);
CREATE INDEX idx_apparats_deleted_at ON apparats(deleted_at);

-- Buildings
CREATE TABLE buildings (
    id UUID PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    iws_code VARCHAR(4),
    building_group INT NOT NULL
);

CREATE INDEX idx_buildings_deleted_at ON buildings(deleted_at);
CREATE INDEX idx_buildings_iws_code ON buildings(iws_code);

-- Control cabinets
CREATE TABLE control_cabinets (
    id UUID PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    building_id UUID NOT NULL,
    project_id UUID,
    control_cabinet_nr VARCHAR(11),
    CONSTRAINT fk_control_cabinets_building FOREIGN KEY (building_id) REFERENCES buildings(id) ON DELETE CASCADE,
    CONSTRAINT fk_control_cabinets_project FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE SET NULL
);

CREATE INDEX idx_control_cabinets_deleted_at ON control_cabinets(deleted_at);
CREATE INDEX idx_control_cabinets_building_id ON control_cabinets(building_id);
CREATE INDEX idx_control_cabinets_project_id ON control_cabinets(project_id);
CREATE INDEX idx_control_cabinets_control_cabinet_nr ON control_cabinets(control_cabinet_nr);

-- SPS controllers
CREATE TABLE sps_controllers (
    id UUID PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    control_cabinet_id UUID NOT NULL,
    project_id UUID,
    ga_device VARCHAR(10),
    device_name VARCHAR(100) NOT NULL,
    device_description VARCHAR(250),
    device_location VARCHAR(250),
    ip_address VARCHAR(50),
    subnet VARCHAR(50),
    gateway VARCHAR(50),
    vlan VARCHAR(50),
    CONSTRAINT fk_sps_controllers_control_cabinet FOREIGN KEY (control_cabinet_id) REFERENCES control_cabinets(id) ON DELETE CASCADE,
    CONSTRAINT fk_sps_controllers_project FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE SET NULL
);

CREATE INDEX idx_sps_controllers_deleted_at ON sps_controllers(deleted_at);
CREATE INDEX idx_sps_controllers_control_cabinet_id ON sps_controllers(control_cabinet_id);
CREATE INDEX idx_sps_controllers_project_id ON sps_controllers(project_id);
CREATE INDEX idx_sps_controllers_ip_address ON sps_controllers(ip_address);
CREATE INDEX idx_sps_controllers_vlan ON sps_controllers(vlan);

-- SPS controller system types
CREATE TABLE sps_controller_system_types (
    id UUID PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    number INT,
    document_name VARCHAR(250),
    sps_controller_id UUID NOT NULL,
    system_type_id UUID NOT NULL,
    CONSTRAINT fk_sps_controller_system_types_sps FOREIGN KEY (sps_controller_id) REFERENCES sps_controllers(id) ON DELETE CASCADE,
    CONSTRAINT fk_sps_controller_system_types_system_type FOREIGN KEY (system_type_id) REFERENCES system_types(id) ON DELETE RESTRICT
);

CREATE INDEX idx_sps_controller_system_types_deleted_at ON sps_controller_system_types(deleted_at);
CREATE INDEX idx_sps_controller_system_types_sps_controller_id ON sps_controller_system_types(sps_controller_id);
CREATE INDEX idx_sps_controller_system_types_system_type_id ON sps_controller_system_types(system_type_id);

-- Field devices
CREATE TABLE field_devices (
    id UUID PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    bmk VARCHAR(10),
    description VARCHAR(250),
    apparat_nr INT,
    sps_controller_system_type_id UUID NOT NULL,
    system_part_id UUID NOT NULL,
    specification_id UUID,
    project_id UUID,
    apparat_id UUID NOT NULL,
    CONSTRAINT fk_field_devices_sps_controller_system_type FOREIGN KEY (sps_controller_system_type_id) REFERENCES sps_controller_system_types(id) ON DELETE RESTRICT,
    CONSTRAINT fk_field_devices_system_part FOREIGN KEY (system_part_id) REFERENCES system_parts(id) ON DELETE CASCADE,
    CONSTRAINT fk_field_devices_specification FOREIGN KEY (specification_id) REFERENCES specifications(id) ON DELETE CASCADE,
    CONSTRAINT fk_field_devices_project FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE SET NULL,
    CONSTRAINT fk_field_devices_apparat FOREIGN KEY (apparat_id) REFERENCES apparats(id) ON DELETE RESTRICT
);

CREATE INDEX idx_field_devices_deleted_at ON field_devices(deleted_at);
CREATE INDEX idx_field_devices_bmk ON field_devices(bmk);
CREATE INDEX idx_field_devices_apparat_nr ON field_devices(apparat_nr);
CREATE INDEX idx_field_devices_sps_controller_system_type_id ON field_devices(sps_controller_system_type_id);
CREATE INDEX idx_field_devices_project_id ON field_devices(project_id);

-- BACnet objects
CREATE TABLE bacnet_objects (
    id UUID PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    text_fix VARCHAR(250) NOT NULL,
    description TEXT,
    gms_visible BOOLEAN NOT NULL DEFAULT FALSE,
    optional BOOLEAN NOT NULL DEFAULT FALSE,
    text_individual VARCHAR(250),
    software_type VARCHAR(2) NOT NULL,
    software_number INT NOT NULL,
    hardware_type VARCHAR(2) NOT NULL,
    hardware_quantity SMALLINT NOT NULL DEFAULT 1,
    field_device_id UUID,
    software_reference_id UUID,
    state_text_id UUID,
    notification_class_id UUID,
    alarm_definition_id UUID,
    CONSTRAINT fk_bacnet_objects_field_device FOREIGN KEY (field_device_id) REFERENCES field_devices(id) ON DELETE CASCADE,
    CONSTRAINT fk_bacnet_objects_software_reference FOREIGN KEY (software_reference_id) REFERENCES bacnet_objects(id) ON DELETE SET NULL,
    CONSTRAINT fk_bacnet_objects_state_text FOREIGN KEY (state_text_id) REFERENCES state_texts(id) ON DELETE SET NULL,
    CONSTRAINT fk_bacnet_objects_notification_class FOREIGN KEY (notification_class_id) REFERENCES notification_classes(id) ON DELETE SET NULL,
    CONSTRAINT fk_bacnet_objects_alarm_definition FOREIGN KEY (alarm_definition_id) REFERENCES alarm_definitions(id) ON DELETE SET NULL
);

CREATE INDEX idx_bacnet_objects_deleted_at ON bacnet_objects(deleted_at);
CREATE UNIQUE INDEX idx_bacnet_fd_text ON bacnet_objects(field_device_id, text_fix);

-- Object data
CREATE TABLE object_data (
    id UUID PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    description VARCHAR(250) NOT NULL,
    version VARCHAR(50) NOT NULL DEFAULT '1.0.0',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    project_id UUID,
    CONSTRAINT fk_object_data_project FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE SET NULL
);

CREATE INDEX idx_object_data_deleted_at ON object_data(deleted_at);
CREATE INDEX idx_object_data_is_active ON object_data(is_active);
CREATE UNIQUE INDEX idx_obj_data_proj_desc ON object_data(project_id, description);

-- Join table: apparat_system_parts
CREATE TABLE apparat_system_parts (
    apparat_id UUID NOT NULL,
    system_part_id UUID NOT NULL,
    PRIMARY KEY (apparat_id, system_part_id),
    CONSTRAINT fk_apparat_system_parts_apparat FOREIGN KEY (apparat_id) REFERENCES apparats(id) ON DELETE CASCADE,
    CONSTRAINT fk_apparat_system_parts_system_part FOREIGN KEY (system_part_id) REFERENCES system_parts(id) ON DELETE CASCADE
);

CREATE INDEX idx_apparat_system_parts_system_part_id ON apparat_system_parts(system_part_id);

-- Join table: object_data_apparats
CREATE TABLE object_data_apparats (
    object_data_id UUID NOT NULL,
    apparat_id UUID NOT NULL,
    PRIMARY KEY (object_data_id, apparat_id),
    CONSTRAINT fk_object_data_apparats_object_data FOREIGN KEY (object_data_id) REFERENCES object_data(id) ON DELETE CASCADE,
    CONSTRAINT fk_object_data_apparats_apparat FOREIGN KEY (apparat_id) REFERENCES apparats(id) ON DELETE CASCADE
);

CREATE INDEX idx_object_data_apparats_apparat_id ON object_data_apparats(apparat_id);

-- Join table: object_data_bacnet_objects
CREATE TABLE object_data_bacnet_objects (
    object_data_id UUID NOT NULL,
    bacnet_object_id UUID NOT NULL,
    PRIMARY KEY (object_data_id, bacnet_object_id),
    CONSTRAINT fk_object_data_bacnet_objects_object_data FOREIGN KEY (object_data_id) REFERENCES object_data(id) ON DELETE CASCADE,
    CONSTRAINT fk_object_data_bacnet_objects_bacnet_object FOREIGN KEY (bacnet_object_id) REFERENCES bacnet_objects(id) ON DELETE CASCADE
);

CREATE INDEX idx_object_data_bacnet_objects_bacnet_object_id ON object_data_bacnet_objects(bacnet_object_id);

-- Object data history
CREATE TABLE object_data_histories (
    id UUID PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    object_data_id UUID,
    user_id UUID,
    action TEXT NOT NULL,
    changes JSON,
    CONSTRAINT fk_object_data_histories_object_data FOREIGN KEY (object_data_id) REFERENCES object_data(id) ON DELETE SET NULL,
    CONSTRAINT fk_object_data_histories_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
);

CREATE INDEX idx_object_data_histories_deleted_at ON object_data_histories(deleted_at);
