/**
 * Field Device domain types
 * Mirrors backend: internal/domain/facility/field_device.go
 */

import type { Pagination } from '../utils/index.ts';

export interface FieldDevice {
	id: string;
	bmk?: string;
	description?: string;
	apparat_nr: string;
	sps_controller_system_type_id: string;
	system_part_id?: string;
	specification_id?: string;
	apparat_id: string;
	created_at: string;
	updated_at: string;
}

export interface BacnetObjectInput {
	text_fix: string;
	description?: string;
	gms_visible: boolean;
	optional: boolean;
	text_individual?: string;
	software_type: string;
	software_number: number;
	hardware_type: string;
	hardware_quantity: number;
	software_reference_id?: string;
	state_text_id?: string;
	notification_class_id?: string;
	alarm_definition_id?: string;
}

export interface CreateFieldDeviceRequest {
	bmk?: string;
	description?: string;
	apparat_nr: number;
	sps_controller_system_type_id: string;
	system_part_id: string;
	apparat_id: string;
	object_data_id?: string;
	bacnet_objects?: BacnetObjectInput[];
}

export interface UpdateFieldDeviceRequest {
	id: string;
	bmk?: string;
	description?: string;
	apparat_nr?: number;
	sps_controller_system_type_id?: string;
	system_part_id?: string;
	apparat_id?: string;
	object_data_id?: string;
	bacnet_objects?: BacnetObjectInput[];
}

export interface FieldDeviceListParams {
	page?: number;
	limit?: number;
	search?: string;
	sps_controller_system_type_id?: string;
}

export interface FieldDeviceListResponse extends Pagination<FieldDevice> {}
