/**
 * Field Device domain types
 * Mirrors backend: internal/domain/facility/field_device.go
 */

import type { Pagination } from '../utils/index.ts';
import type { Apparat } from './apparat.js';
import type { BacnetObject } from './bacnet-object.js';
import type { Specification } from './specification.js';
import type { SPSControllerSystemType } from './sps-sys-type.js';
import type { SystemPart } from './system.js';

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

	// Embedded related entities for display
	sps_controller_system_type?: SPSControllerSystemType;
	apparat?: Apparat;
	system_part?: SystemPart;
	specification?: Specification;
	bacnet_objects?: BacnetObject[];
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

/**
 * FieldDeviceOptions - Response from /api/v1/facility/field-devices/options
 * Contains all metadata needed for creating/editing field devices with relationships
 */
export interface FieldDeviceOptions {
	apparats: import('./apparat.js').Apparat[];
	system_parts: import('./system.js').SystemPart[];
	object_datas: import('./object-data.js').ObjectData[];
	apparat_to_system_part: Record<string, string[]>; // apparat_id -> [system_part_ids]
	object_data_to_apparat: Record<string, string[]>; // object_data_id -> [apparat_ids]
}

/**
 * Available apparat numbers response
 */
export interface AvailableApparatNumbersResponse {
	available: number[];
}

/**
 * Multi-create field device request
 */
export interface MultiCreateFieldDeviceRequest {
	field_devices: CreateFieldDeviceRequest[];
}

/**
 * Result for a single field device creation in multi-create
 */
export interface FieldDeviceCreateResult {
	index: number; // Index in the original request array
	success: boolean; // Whether the creation succeeded
	field_device?: FieldDevice; // The created field device (null if failed)
	error: string; // Error message if failed (empty if succeeded)
	error_field: string; // Specific field that caused the error (if applicable)
}

/**
 * Multi-create field device response
 */
export interface MultiCreateFieldDeviceResponse {
	results: FieldDeviceCreateResult[];
	total_requests: number;
	success_count: number;
	failure_count: number;
}

/**
 * Bulk update field device item
 */
export interface BulkUpdateFieldDeviceItem {
	id: string;
	bmk?: string;
	description?: string;
	apparat_nr?: number;
	apparat_id?: string;
	system_part_id?: string;
}

/**
 * Bulk update field device request
 */
export interface BulkUpdateFieldDeviceRequest {
	updates: BulkUpdateFieldDeviceItem[];
}

/**
 * Bulk operation result item
 */
export interface BulkOperationResultItem {
	id: string;
	success: boolean;
	error?: string;
}

/**
 * Bulk update field device response
 */
export interface BulkUpdateFieldDeviceResponse {
	results: BulkOperationResultItem[];
	total_count: number;
	success_count: number;
	failure_count: number;
}

/**
 * Bulk delete field device request
 */
export interface BulkDeleteFieldDeviceRequest {
	ids: string[];
}

/**
 * Bulk delete field device response
 */
export interface BulkDeleteFieldDeviceResponse {
	results: BulkOperationResultItem[];
	total_count: number;
	success_count: number;
	failure_count: number;
}
